#!/usr/bin/env python3
"""Command-line client for the Kkal Tracker food API (standard library only).

Commands:
  ingredients              list the user's ingredients (compact table, most used first)
  add                      store one meal; the meal JSON comes from --json or stdin
  entries [--from --to]    list diary entries with their IDs (default: today, UTC)
  delete ENTRY_ID          delete a diary entry

Every command prints the HTTP status on failure and exits with code 1.
"""

import argparse
import datetime
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request

# --- Configuration -----------------------------------------------------------
# Create the key in Kkal Tracker: Settings -> API Keys. Put it into a file named "api_key"
# in the skill folder (next to SKILL.md) or paste it here.
# The environment variables KKAL_API_KEY and KKAL_BASE_URL take precedence.
DEFAULT_BASE_URL = "http://37.27.186.57:8080"
API_KEY = "XXXXXXXXX"
# -----------------------------------------------------------------------------

API_KEY_PLACEHOLDER = "XXXXXXXXX"
API_KEY_FILE = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "api_key")
TIMEOUT_SECONDS = 30


def load_api_key():
    """Environment variable, then the api_key file of the skill, then the constant above."""
    key = os.environ.get("KKAL_API_KEY", "").strip()
    if key:
        return key
    try:
        with open(API_KEY_FILE, encoding="utf-8") as key_file:
            key = key_file.read().strip()
    except OSError:
        key = ""
    return key or API_KEY


BASE_URL = os.environ.get("KKAL_BASE_URL", DEFAULT_BASE_URL).rstrip("/")
API_KEY = load_api_key()


def fail(message):
    print(message, file=sys.stderr)
    sys.exit(1)


def request(method, path, body=None):
    """Send a request and return (status, parsed JSON or None)."""
    if not API_KEY or API_KEY == API_KEY_PLACEHOLDER:
        fail("API key is not configured. Create one in Kkal Tracker (Settings -> API Keys) "
             f"and save it to {API_KEY_FILE} or set KKAL_API_KEY.")

    headers = {"X-API-Key": API_KEY, "Accept": "application/json"}
    data = None
    if body is not None:
        data = json.dumps(body).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(BASE_URL + "/api/v1" + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT_SECONDS) as response:
            raw = response.read()
            return response.status, json.loads(raw) if raw else None
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


def cmd_delete(args):
    status, payload = request("DELETE", f"/food/{args.entry_id}")
    expect(status, payload, 204)
    print(f"Entry {args.entry_id} deleted")


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

    delete = commands.add_parser("delete", help="delete a diary entry")
    delete.add_argument("entry_id", type=int)
    delete.set_defaults(func=cmd_delete)

    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
