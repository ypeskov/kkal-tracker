#!/usr/bin/env python3
"""Command-line client for the Kkal Tracker food API (standard library only).

Commands:
  ingredients              list the user's ingredients (compact table, most used first)
  add                      store one meal; the meal JSON comes from --json or stdin
  entries [--from --to]    list diary entries with their IDs (default: today, UTC)
  totals --day YYYY-MM-DD  calories, fats, carbs and proteins of a meal (--meal IDs) and of the whole diary day
  edit ENTRY_ID [...]      change the weight, ingredient, calories, macros or time of a diary entry
  delete ENTRY_ID          delete a diary entry
  edit-ingredient ID [...] correct the name, calories or nutrients of an ingredient in the list
  delete-ingredient ID     delete an ingredient from the list (diary entries stay)

Every command prints the HTTP status on failure and exits with code 1.
"""

import argparse
import datetime
import decimal
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request

# --- Configuration -----------------------------------------------------------
# The skill downloaded from Kkal Tracker (Settings -> AI -> Download skill) is ready to use: the download puts
# the server address and a new API key of the user into the two lines below. If the key is not configured,
# download the skill again and install it in place of this one.
# For manual checks the environment variables KKAL_API_KEY and KKAL_BASE_URL override these values.
DEFAULT_BASE_URL = "https://kcal.peskov.info"
API_KEY = "API_KEY_HERE"
# -----------------------------------------------------------------------------

# Assembled from parts: the download replaces the only literal placeholder, the one in the line above
API_KEY_PLACEHOLDER = "API_KEY_" + "HERE"
TIMEOUT_SECONDS = 30
NUTRIENTS = ("fats", "carbs", "proteins")


def load_key():
    """The KKAL_API_KEY environment variable, then the constant above."""
    return os.environ.get("KKAL_API_KEY", "").strip() or API_KEY


BASE_URL = os.environ.get("KKAL_BASE_URL", DEFAULT_BASE_URL).rstrip("/")
API_KEY = load_key()


def fail(message):
    print(message, file=sys.stderr)
    sys.exit(1)


def request(method, path, body=None):
    """Send a request and return (status, parsed JSON or None)."""
    if not API_KEY or API_KEY == API_KEY_PLACEHOLDER:
        fail("API key is not configured. Download the skill again in Kkal Tracker (Settings -> AI -> "
             "Download skill) and install it in place of this one.")

    headers = {"X-API-Key": API_KEY, "Accept": "application/json"}
    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(BASE_URL + "/api/v1" + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT_SECONDS) as response:
            raw = response.read()
            try:
                return response.status, json.loads(raw) if raw else None
            except ValueError:
                # An older server answers unknown API calls with the website page instead of an error
                fail(f"HTTP {response.status}, but the answer is not JSON: the server does not support "
                     f"{method} /api/v1{path.split('?')[0]} (probably an older server version). Nothing was changed.")
    except urllib.error.HTTPError as error:
        raw = error.read()
        try:
            payload = json.loads(raw)
        except ValueError:
            payload = {"message": raw.decode("utf-8", "replace")}
        return error.code, payload
    except (urllib.error.URLError, TimeoutError) as error:
        # The request may or may not have reached the server: the caller must check before retrying
        fail(f"NETWORK ERROR: {error}. For 'add', run 'entries' to see whether the meal was stored before retrying.")


def print_json(payload):
    print(json.dumps(payload, ensure_ascii=False, indent=2))


def expect(status, payload, ok_status):
    if status != ok_status:
        print(f"HTTP {status}", file=sys.stderr)
        if payload is not None:
            print_json(payload)
        sys.exit(1)


def number(value):
    """Format a number without a trailing .0; missing values become '-'."""
    if value is None:
        return "-"
    return f"{value:g}"


def expect_ingredient_change(status, payload, ok_status):
    """Like expect, but tells a missing ingredient from an older server without the ingredient routes."""
    not_found = isinstance(payload, dict) and payload.get("message") == "Ingredient not found"
    if status in (404, 405) and not not_found:
        fail(f"HTTP {status}: the server probably does not support editing or deleting ingredients yet "
             "(an older server version). Nothing was changed.")
    expect(status, payload, ok_status)


def cmd_ingredients(args):
    status, payload = request("GET", "/ingredients")
    expect(status, payload, 200)
    if args.json:
        print_json(payload)
        return

    ingredients = sorted(payload["ingredients"], key=lambda i: (-i["times_used"], i["name"].lower()))
    print("id\tname\tkcal_per_100g\tfats\tcarbs\tproteins\ttimes_used\tlast_used")
    for i in ingredients:
        print("\t".join([
            str(i["id"]), i["name"], number(i["kcal_per_100g"]),
            number(i.get("fats")), number(i.get("carbs")), number(i.get("proteins")),
            str(i["times_used"]), i.get("last_used") or "-",
        ]))


def cmd_add(args):
    raw = args.json if args.json is not None else sys.stdin.read()
    try:
        meal = json.loads(raw)
    except ValueError as error:
        fail(f"The meal is not valid JSON: {error}")

    status, payload = request("POST", "/food", meal)
    expect(status, payload, 201)
    print_json(payload)


def cmd_entries(args):
    today = datetime.datetime.now(datetime.timezone.utc).date().isoformat()
    query = urllib.parse.urlencode({"type": "food", "from": args.date_from or today, "to": args.date_to or args.date_from or today})
    status, payload = request("GET", "/data?" + query)
    expect(status, payload, 200)
    print_json(payload.get("food", []))


def grams(value):
    """Round grams to a whole number, halves up (Python's round() would round 12.5 down to 12)."""
    return int(value.quantize(decimal.Decimal(1), rounding=decimal.ROUND_HALF_UP))


def totals_row(entries):
    """Calories (the server's entry calories) and grams of fats, carbs and proteins of the entries, unrounded."""
    row = {"kcal": sum(entry["calories"] for entry in entries)}
    for nutrient in NUTRIENTS:
        # Decimal from the shortest text form of each number: no binary float error before the halves-up rounding
        row[nutrient] = sum((decimal.Decimal(str(entry.get(nutrient) or 0)) * decimal.Decimal(str(entry["weight"]))
                             / 100 for entry in entries), decimal.Decimal(0))
    return row


def rounded(row):
    return {field: grams(value) if field in NUTRIENTS else value for field, value in row.items()}


def cmd_totals(args):
    day = args.day
    try:
        valid_day = len(day) == 10 and datetime.datetime.strptime(day, "%Y-%m-%d").date().isoformat() == day
    except ValueError:
        valid_day = False
    if not valid_day:
        fail(f"Invalid --day {day!r}: expected a date as YYYY-MM-DD, e.g. the 'day' of the add or edit response")
    if args.expect_meal_kcal is not None and not args.meal:
        fail("--expect-meal-kcal needs the entry IDs of the meal in --meal")

    query = urllib.parse.urlencode({"type": "food", "from": day, "to": day})
    status, payload = request("GET", "/data?" + query)
    expect(status, payload, 200)
    entries = (payload or {}).get("food") or []

    meal_ids = list(dict.fromkeys(args.meal or []))
    found_ids = {entry["id"] for entry in entries}
    meal = None
    if args.meal:
        meal_entries = [entry for entry in entries if entry["id"] in meal_ids]
        meal = {"entry_ids": [entry_id for entry_id in meal_ids if entry_id in found_ids]}
        meal.update(rounded(totals_row(meal_entries)))
    day_total = {"entries": len(entries)}
    day_total.update(rounded(totals_row(entries)))
    missing_meal_ids = [entry_id for entry_id in meal_ids if entry_id not in found_ids]

    check = {"meal": None, "day": None}
    if args.expect_meal_kcal is not None:
        meal_ok = meal["kcal"] == args.expect_meal_kcal and not missing_meal_ids
        check["meal"] = "ok" if meal_ok else "mismatch"
    if args.expect_day_kcal is not None:
        check["day"] = "ok" if day_total["kcal"] == args.expect_day_kcal else "mismatch"

    print_json({
        "day": day,
        "meal": meal,
        "day_total": day_total,
        "missing_meal_ids": missing_meal_ids,
        "without_macros": [
            {"id": entry["id"], "food": entry["food"], "in_meal": entry["id"] in meal_ids}
            for entry in entries if all(entry.get(nutrient) is None for nutrient in NUTRIENTS)
        ],
        "check": check,
    })


def cmd_edit(args):
    changes = {
        "weight": args.weight,
        "ingredient_id": args.ingredient_id,
        "kcal_per_100g": args.kcal,
        "fats": args.fats,
        "carbs": args.carbs,
        "proteins": args.proteins,
        "meal_datetime": args.datetime,
    }
    changes = {field: value for field, value in changes.items() if value is not None}
    if not changes:
        fail("Nothing to change: pass --weight, --ingredient-id, --kcal, --fats, --carbs, --proteins or --datetime")

    status, payload = request("PUT", f"/food/{args.entry_id}", changes)
    expect(status, payload, 200)
    print_json(payload)


def cmd_delete(args):
    status, payload = request("DELETE", f"/food/{args.entry_id}")
    expect(status, payload, 204)
    print(f"Entry {args.entry_id} deleted")


def cmd_edit_ingredient(args):
    changes = {
        "name": args.name,
        "kcal_per_100g": args.kcal,
        "fats": args.fats,
        "carbs": args.carbs,
        "proteins": args.proteins,
    }
    changes = {field: value for field, value in changes.items() if value is not None}
    if not changes:
        fail("Nothing to change: pass --name, --kcal, --fats, --carbs or --proteins")

    status, payload = request("PUT", f"/ingredients/{args.ingredient_id}", changes)
    expect_ingredient_change(status, payload, 200)
    print_json(payload)


def cmd_delete_ingredient(args):
    status, payload = request("DELETE", f"/ingredients/{args.ingredient_id}")
    expect_ingredient_change(status, payload, 204)
    print(f"Ingredient {args.ingredient_id} deleted")


def main():
    parser = argparse.ArgumentParser(description="Kkal Tracker food API client")
    commands = parser.add_subparsers(dest="command", required=True)

    ingredients = commands.add_parser("ingredients", help="list ingredients")
    ingredients.add_argument("--json", action="store_true", help="print the raw JSON response")
    ingredients.set_defaults(func=cmd_ingredients)

    add = commands.add_parser("add", help="store one meal")
    add.add_argument("--json", help="meal as a JSON string (read from stdin when omitted)")
    add.set_defaults(func=cmd_add)

    entries = commands.add_parser("entries", help="list diary entries with IDs")
    entries.add_argument("--from", dest="date_from", help="YYYY-MM-DD (default: today, UTC)")
    entries.add_argument("--to", dest="date_to", help="YYYY-MM-DD (default: same as --from)")
    entries.set_defaults(func=cmd_entries)

    totals = commands.add_parser(
        "totals", help="calories and grams of fats, carbs and proteins of a meal and of the whole diary day")
    totals.add_argument("--day", required=True, metavar="YYYY-MM-DD",
                        help="the diary day, YYYY-MM-DD: the 'day' of the add or edit response (UTC)")
    totals.add_argument("--meal", type=int, nargs="+", action="extend", metavar="ID",
                        help="entry IDs of the meal just stored; without it the meal row is null")
    totals.add_argument("--expect-meal-kcal", type=int, metavar="N",
                        help="total_calories of the add response: check.meal is ok when the meal adds up to it")
    totals.add_argument("--expect-day-kcal", type=int, metavar="N",
                        help="day_total_calories of the add or edit response: check.day is ok when the day adds up to it")
    totals.set_defaults(func=cmd_totals)

    edit = commands.add_parser("edit", help="change a diary entry; options that are not passed keep their values")
    edit.add_argument("entry_id", type=int)
    edit.add_argument("--weight", type=float, help="new weight in grams")
    edit.add_argument("--ingredient-id", type=int, help="switch the entry to another existing ingredient")
    edit.add_argument("--kcal", type=float, help="calories per 100 g for this entry only")
    edit.add_argument("--fats", type=float, help="fats, grams per 100 g, for this entry only")
    edit.add_argument("--carbs", type=float, help="carbohydrates, grams per 100 g, for this entry only")
    edit.add_argument("--proteins", type=float, help="proteins, grams per 100 g, for this entry only")
    edit.add_argument("--datetime", help="new meal time, RFC 3339 with the UTC offset")
    edit.set_defaults(func=cmd_edit)

    delete = commands.add_parser("delete", help="delete a diary entry")
    delete.add_argument("entry_id", type=int)
    delete.set_defaults(func=cmd_delete)

    edit_ingredient = commands.add_parser(
        "edit-ingredient",
        help="correct an ingredient of the list for all future entries; logged entries do not change")
    edit_ingredient.add_argument("ingredient_id", type=int)
    edit_ingredient.add_argument("--name", help="new name")
    edit_ingredient.add_argument("--kcal", type=float,
                                 help="calories per 100 g of the ingredient, for all future entries "
                                      "(unlike 'edit --kcal', which changes one entry only)")
    edit_ingredient.add_argument("--fats", type=float, help="fats, grams per 100 g")
    edit_ingredient.add_argument("--carbs", type=float, help="carbohydrates, grams per 100 g")
    edit_ingredient.add_argument("--proteins", type=float, help="proteins, grams per 100 g")
    edit_ingredient.set_defaults(func=cmd_edit_ingredient)

    delete_ingredient = commands.add_parser(
        "delete-ingredient", help="delete an ingredient from the list; logged entries stay, no undo")
    delete_ingredient.add_argument("ingredient_id", type=int)
    delete_ingredient.set_defaults(func=cmd_delete_ingredient)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
