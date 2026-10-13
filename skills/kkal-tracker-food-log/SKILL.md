---
name: kkal-tracker-food-log
description: Record what the user ate or drank into their Kkal Tracker food diary (kcal.peskov.info) through its API. Use it whenever the user tells you what they ate, dictates a meal or snack, lists foods with weights, or asks to log, add or save food or calories, including casual or voice-dictated phrasings like "я съел 200 грамм курицы и помидор", "запиши обед", "добавь в дневник", "I had two eggs and toast". Also use it to correct or delete a recorded entry, to fix or delete an ingredient in the user's list, and to check what is logged today. The user does not need to name Kkal Tracker. This is the default for food logging: the "-dev" copy is only for a message that itself names the dev or test tracker. Edit or delete an entry or ingredient with the copy that created it.
---

# Kkal Tracker: food diary logging

This skill writes meals into the user's Kkal Tracker diary. The user usually dictates by voice: one phrase may contain several foods, the names are approximate, and weights may be missing. Your job is to turn that phrase into exact diary entries without making the user repeat themselves.

The server is deliberately strict: it never guesses. Every food you send is either an ingredient the user already has (referenced by `ingredient_id`) or an explicitly described new food. Choosing between the two is your part of the work.

## Production and dev: where a request belongs

Two copies of this skill can be installed: the regular one writes to the real diary, the `-dev` one to a test server. A mix-up either puts test data into the real diary or loses a real meal in the test database, so the routing is strict.

- **Logging food** — every message is routed on its own. It goes to dev only when that message itself names the dev or test tracker ("запиши на дев", "в тестовый трекер", "log it to dev"). A dev request earlier in the conversation does not carry over: a following "добавь ещё кофе" without the word goes to the real diary, even in the middle of a testing session.
- **Editing and deleting** — these follow the entry, not the wording. Use the copy that created the entry, however the request is phrased; replies of the dev copy start with `[DEV]`, so the conversation shows where each entry went. Entry IDs are separate on each server and the same number can belong to a different, real entry on the other one, so an edit or delete sent to the wrong server silently damages an unrelated entry.
- **Ingredients** follow the same rule: ingredient IDs are separate on each server too. Edit or delete an ingredient with the copy whose `ingredients` list it came from.
- If you cannot tell where an entry or an ingredient came from, find it with `entries` or `ingredients` before changing anything, and ask the user when it is still unclear.

## API basics

- Base URL: `https://kcal.peskov.info`
- Auth: `X-API-Key` header. The skill downloaded from Kkal Tracker (Settings → AI → Download skill) already contains its own key in `scripts/kkal_client.py`, nothing has to be configured. The `KKAL_API_KEY` environment variable overrides it, for manual checks only. If the script says the key is not configured, or a request returns `401`, ask the user to download the skill again in Kkal Tracker (Settings → AI) and install it in place of this one.
- The key is a secret: never print, quote or show the `API_KEY` line or the key itself, and do not open `scripts/kkal_client.py` to read it. Just run the client.
- Rate limit: a burst of 20 requests, then 1 request per second. A normal logging takes three calls: `ingredients`, `add` and `totals`.

Use the bundled client, it handles the key, headers and errors:

```bash
python scripts/kkal_client.py ingredients            # the user's ingredient list
python scripts/kkal_client.py add --json '<meal>'    # store a meal (or pipe the JSON to stdin)
python scripts/kkal_client.py entries                # today's diary entries with IDs
python scripts/kkal_client.py entries --from 2026-03-01 --to 2026-03-07
python scripts/kkal_client.py totals --day 2026-10-13 --meal 3622 3623 3624 --expect-meal-kcal 888 --expect-day-kcal 1640   # meal and day macros
python scripts/kkal_client.py edit 12345 --weight 150   # correct an entry
python scripts/kkal_client.py edit 12345 --kcal 260 --fats 3 --carbs 48 --proteins 9   # label values for this entry only
python scripts/kkal_client.py delete 12345           # delete an entry
python scripts/kkal_client.py edit-ingredient 94 --kcal 260   # correct an ingredient of the list
python scripts/kkal_client.py delete-ingredient 94   # delete an ingredient of the list
```

## Workflow

### 1. Fetch the ingredient list

Run `ingredients` before logging. The user keeps adding foods through the website, so fetch the list fresh at the start of every conversation rather than relying on memory. Within one conversation the list you already have stays valid — the `add` response gives you the IDs of anything you create yourself.

The output is a tab-separated table sorted by how often the user logs each food:

```
id   name                      kcal_per_100g  fats  carbs  proteins  times_used  last_used
75   Помидор                   20             0.2   4.2    0.6       314         2026-10-02
120  Свиная шея (вес сырой)    267            22.5  -      16        61          2026-09-30
```

### 2. Split the phrase into foods and weights

Each food becomes one item with a weight **in grams**.

- Convert units: "полкило" → 500, "стакан кефира" → about 250, "столовая ложка масла" → about 15.
- When the user gives pieces instead of grams ("помидор", "два яйца", "кусок хлеба"), use a typical weight for that food and say which weight you assumed in your reply.
- If the amount is missing entirely and there is no sensible default (e.g. "поел гречки"), ask — a guessed portion can be off by a factor of three.

### 3. Match each food to an ingredient

Match by meaning, not by string equality: the user says "шея", the list has "Свиная шея (вес сырой)". Names in the list may be in several languages and often carry qualifiers that matter:

- **State and preparation** — "(вес сырой)", "(уже вареная)", "жареная". Raw and cooked versions differ a lot in calories per 100 g. If the user did not say which one they mean and the list has only one, use it. If the list has both, go by what they described ("сварил 100 грамм гречки" is dry weight; "съел тарелку гречки" is cooked).
- **Brand or variety** — "Сыр Ементаль" vs "Сыр Ементаль плавленный", "Сръбска наденица" vs "Сръбска наденица тънка".
- **Usage statistics** — when the phrase is generic ("сыр", "колбаса") and several ingredients fit, `times_used` and `last_used` show the user's habit. A clear favourite (used far more often or just recently) is almost certainly what they mean.

Decide like this:

- **One clear match** → use its `ingredient_id`.
- **Several plausible matches with similar calories** → take the most used one and mention the choice in the reply.
- **Several plausible matches with noticeably different calories and no clear favourite** → ask one short question listing the candidates. A wrong pick here silently corrupts the diary.
- **No match** → it is a new food, see the next step.

If the user states a calorie value for a known food that differs from the list ("этот хлеб 260 ккал"), keep the `ingredient_id` and add `kcal_per_100g` to the item: it overrides the value for this entry only. When they read the label values ("на упаковке 260 ккал, Ж 3, У 48, Б 9"), override the calories and the macros together: `kcal_per_100g`, `fats`, `carbs` and `proteins` (grams per 100 g) on the item. Each value you send replaces the ingredient's value for this entry only; the ones you do not send come from the ingredient. If the user gives only the calories, override only `kcal_per_100g` and say in the reply that the macros still come from the ingredient. Offer in one line of the reply to update the value in the ingredient list as well, and change the list only after the user agrees (see "Correcting the ingredient list"). When the user points at the product itself ("у этого хлеба на самом деле 260 ккал", "он всегда 260"), correct the ingredient first with `edit-ingredient` and then log it without an override.

### 4. New foods

When nothing in the list fits, create the food instead of interrogating the user:

- Use the numbers the user gave (from a package, a menu) when there are any.
- Otherwise estimate calories, fats, carbs and proteins **per 100 g** from general nutrition knowledge.
- Name it short and the way the user calls it, in the user's language. It becomes a permanent entry in their list, so "Сырники" is better than "сырники домашние которые я ел сегодня".
- Always tell the user in the reply that a new ingredient was created and with which values, so they can correct a bad estimate.

**One-off foods.** Some foods will never be logged again: a restaurant dish, food at a party, something bought on a trip. Saving each of them would clutter the ingredient list and make future matching harder. Set `"one_off": true` on such items — the entry goes to the diary and counts toward calories, but nothing is added to the ingredient list. Use it when the user says so ("разово", "не сохраняй") or when the context makes it obvious (restaurant, guests, travel). When unsure, save the ingredient: an ordinary product is likely to come up again.

If the user only knows the total calories ("ужин в ресторане примерно на 800 ккал"), estimate the portion weight and derive calories per 100 g (`total × 100 / weight`). Do not put the total into `kcal_per_100g` — the server rejects values above 900.

### 5. Send the meal in one request

```json
{
  "meal_datetime": "2026-10-02T13:30:00+03:00",
  "items": [
    {"ingredient_id": 120, "weight": 200},
    {"ingredient_id": 94, "weight": 60, "kcal_per_100g": 260, "fats": 3, "carbs": 48, "proteins": 9},
    {"new_ingredient": {"name": "Сырники", "kcal_per_100g": 220, "fats": 9, "carbs": 20, "proteins": 14}, "weight": 150},
    {"new_ingredient": {"name": "Паста карбонара", "kcal_per_100g": 180, "fats": 9, "carbs": 16, "proteins": 8}, "weight": 300, "one_off": true}
  ]
}
```

- Send all foods of the meal together. The request is atomic: if one item is rejected, nothing is saved, so you can fix the item and resend the whole meal without creating duplicates.
- `meal_datetime` is optional. Omit it when the user is logging what they just ate. Set it (with the user's UTC offset) when they name a time: "на завтрак в 8", "вчера вечером".
- The server calculates the calories itself from weight and calories per 100 g — do not send totals.
- Limits: weight up to 10000 g, `kcal_per_100g` 0–900, macros 0–100 g per 100 g (in `new_ingredient` and in the per-entry overrides `fats`, `carbs`, `proteins`), up to 50 items. The same calorie and macro limits apply to `edit` and `edit-ingredient`.

The response lists the stored entries (with `id`, the resolved `food` name, `calories` and the nutrients `fats`, `carbs`, `proteins` per 100 g when they are known), `total_calories` of this meal, the diary `day` the meal was counted for and `day_total_calories` for that day.

### 6. Add up the macros

The user wants fats, carbs and proteins next to the calories, for this meal and for the whole day. The server totals only calories, so the grams come from the diary entries. Never add up macros by hand or with your own code: run `totals`, it fetches the day and does the sums.

```bash
python scripts/kkal_client.py totals --day 2026-10-13 --meal 3622 3623 3624 --expect-meal-kcal 888 --expect-day-kcal 1640
```

- `--day` is the `day` of the `add` response, `--meal` the `id`s of the entries just stored, `--expect-meal-kcal` the `total_calories` and `--expect-day-kcal` the `day_total_calories` of the response. If the response has no `day_total_calories`, omit `--expect-day-kcal`; the day calories then come from `totals`.
- Always pass the `day` from the response, never "today". It is the diary day the server counted `day_total_calories` for (the diary keeps days as UTC dates). It differs from today when `meal_datetime` names another day ("вчера вечером"), and UTC midnight can pass between `add` and `totals`; neither `entries` nor `totals` can guess it.

The output:

```json
{
  "day": "2026-10-13",
  "meal": {"entry_ids": [3622, 3623, 3624], "kcal": 888, "fats": 59, "carbs": 35, "proteins": 54},
  "day_total": {"entries": 7, "kcal": 1640, "fats": 87, "carbs": 96, "proteins": 112},
  "missing_meal_ids": [],
  "without_macros": [{"id": 3601, "food": "Кофе", "in_meal": false}],
  "check": {"meal": "ok", "day": "ok"}
}
```

- Grams are rounded to whole numbers; a nutrient that is not filled in counts as 0.
- `check.meal` is `"mismatch"` when the meal does not add up to `total_calories` or some of its entries were not found on that day (`missing_meal_ids`). Then show the meal's calories from the server and say that the meal macros could not be calculated.
- `check.day` is `"mismatch"` when the day does not add up to `day_total_calories`. Then show the server's `day_total_calories` and say that the day macros could not be calculated. Never show numbers that do not belong together.
- `without_macros` lists the foods with no fats, carbs or proteins at all (`in_meal` tells whether they belong to this meal). The totals are understated by them, so name them in the reply.

For questions like "сколько за сегодня?" run `totals --day <today's UTC date, the day the website shows>` without `--meal`: `meal` is then `null`, reply with the day row only.

### 7. Reply to the user

The user is often on voice and wants a quick confirmation, not a report. Reply in the user's language, in a few lines:

1. What was recorded: a short list with food, grams and calories for each item.
2. A table with this meal and the whole day. Write it as a real markdown table, **not inside a code block**, so that it renders as a table in the chat.
   - Rows: "Приём" (this meal) and "За день" (the whole diary day).
   - Columns, always in this order: Ккал, Ж, У, Б, that is calories, fats, carbs, proteins (К-Ж-У-Б). This is deliberately not the usual Б-Ж-У order: keep К-Ж-У-Б in every language, with the headers in the user's language.
   - The calories are the server's `total_calories` and `day_total_calories`, the grams come from `totals`, whole numbers.
   - A row whose check is `"mismatch"` shows the server's calories and `—` in the gram cells.
3. Notes, so the user can catch a mistake: weights you assumed, which ingredient you picked when there was a choice, new ingredients created (with the values you estimated), one-off items, foods without macros, and any per-entry override. When only the calories were overridden, say that the macros still come from the ingredient.

Example. The list is shown in a code block only as an illustration; the table is written exactly as it should appear in the reply:

```
Записал:
• Свиная шея (вес сырой) — 200 г, 534 ккал
• Помидор — 120 г, 24 ккал (вес одного помидора принял за 120 г)
• Сырники — 150 г, 330 ккал — новый продукт в справочнике: 220 ккал/100 г, Ж 9 / У 20 / Б 14 (оценка)
```

|         | Ккал  | Ж, г | У, г | Б, г |
|---------|-------|------|------|------|
| Приём   | 888   | 59   | 35   | 54   |
| За день | 1 640 | 87   | 96   | 112  |

## Fixing mistakes

Use the entry `id` from the `add` response. If you no longer have it (another conversation, an entry made on the website), find it with `entries`; when several entries could be the one the user means, ask rather than guess.

- **Correct an entry** ("исправь вес на 150", "там было не 200, а 120 грамм", "это была не шея, а грудка", "это было в обед, а не утром") → `edit`. It changes only what you pass and keeps the rest, including the meal time; the server recalculates the calories and returns the new day total.

  ```bash
  python scripts/kkal_client.py edit 12345 --weight 150
  python scripts/kkal_client.py edit 12345 --ingredient-id 76          # another food from the ingredient list
  python scripts/kkal_client.py edit 12345 --kcal 260                  # calories per 100 g for this entry only
  python scripts/kkal_client.py edit 12345 --kcal 260 --fats 3 --carbs 48 --proteins 9   # label values for this entry only
  python scripts/kkal_client.py edit 12345 --datetime 2026-10-02T13:30:00+03:00
  ```

  The options can be combined. `--ingredient-id` takes the name, calories and nutrients from that ingredient and keeps the weight; `--kcal`, `--fats`, `--carbs` and `--proteins` passed together with it are applied on top. A macro cannot be reset to "unknown"; send `0` if it really is zero.
- **The right food is not in the ingredient list** → `edit` cannot create ingredients. Delete the entry and `add` it again as a new food, passing the original `meal_datetime` so the entry does not move to the current time.
- **Undo** ("отмени", "удали последнее", "это было не то") → `delete` each entry `id`. Undoing an entry never deletes its ingredient, even one you have just created.
- **An ingredient created by mistake** (a duplicate, a food that should have been one-off) → `delete-ingredient` when the user asks for it; **a wrong estimate or a typo** in an ingredient → `edit-ingredient`. See the next section.

After a correction, show the entry as it is now and the "За день" row again, in the table of step 7:

- after `edit`: `totals --day <day of the edit response> --expect-day-kcal <its day_total_calories>`; the calories of the day row come from the edit response;
- after `delete` the server returns nothing, so `totals --day <the entry's day>` is the only source of the day row. The entry's day is the `day` of the `add` or `edit` response that produced it, or the UTC date of its `meal_datetime` in `entries`.

### Correcting the ingredient list

Diary entries keep their own copy of the food name, calories and macros. Changing or deleting an ingredient never changes entries that are already in the diary; only entries logged afterwards use the new values.

**Edit the ingredient or override one entry.**

- **Edit the ingredient** (`edit-ingredient`) only when the user points at the product or the list: "в справочнике", "на самом деле", "всегда" ("у этого продукта на самом деле 260 ккал", "в справочнике неправильно", "он всегда 260"), or asks to fix a typo in a name or a bad estimate of an ingredient you created.
- **Otherwise override the values for the entry only** — `kcal_per_100g`, `fats`, `carbs`, `proteins` on the item, or `edit --kcal … --fats … --carbs … --proteins …` — as before, e.g. "этот хлеб 260 ккал" or label values read while logging. Offer in one line of the reply to update the list as well, and change the list only after the user agrees.

```bash
python scripts/kkal_client.py edit-ingredient 94 --kcal 260
python scripts/kkal_client.py edit-ingredient 94 --name "Хлеб ржаной" --fats 1.2 --carbs 48 --proteins 8.5
```

`edit-ingredient` changes only the options you pass and keeps the rest. A macro cannot be reset to "unknown"; send `0` if it really is zero.

- **Logged entries.** Refresh only the entries of the current conversation that the correction is clearly about (e.g. the meal you just logged with the wrong value): `edit ENTRY_ID --ingredient-id <the same ingredient id>` copies the corrected name, calories and macros into the entry and keeps its weight and time. Leave older entries as they are unless the user asks to update them.
- **Renaming** resets `times_used` and `last_used` of the ingredient (the statistics are matched by name, and old entries keep the old name), so you lose the habit signal used for matching. Rename only when the user asks or the name is clearly wrong.
- **Reply** with the change, old → new values, and the entries you refreshed, if any.

**Delete an ingredient only on the user's explicit request** ("удали из справочника", "убери этот продукт из списка"). Never delete one as a side effect of undoing an entry or to tidy up the list on your own initiative. Tell the user that diary entries with this food stay as they are and still count, and that deletion cannot be undone through the API (the food would have to be created again, with a new ID).

```bash
python scripts/kkal_client.py delete-ingredient 94
```

## Error handling

- `409` on `add` → a new ingredient has the same name as an existing one (case, spacing and "ё/е" are ignored). The response contains `existing_ingredient` and `item_index`: replace that item's `new_ingredient` with the `ingredient_id` and resend the whole meal. Nothing was saved.
- `409` on `edit-ingredient` → another ingredient already has that name; `existing_ingredient` shows it. Nothing was changed. Tell the user and ask whether to keep both under different names or to use the existing one; do not delete or merge anything on your own.
- `400` on `add` → the message names the item (`items[2]: …`) and the problem. Fix it and resend the whole meal. Nothing was saved.
- `400` on `edit` or `edit-ingredient` → the message names the field and the limit. Nothing was changed.
- `404` on `edit` or `delete` → there is no such entry (already deleted, or a wrong ID). Check with `entries`.
- `404` on `edit-ingredient` or `delete-ingredient` → there is no such ingredient (already deleted, a wrong ID, or an ID from the other server). Fetch `ingredients` again.
- The client says the server probably does not support editing or deleting ingredients yet → the server is an older version. Tell the user, do not retry; nothing was changed.
- `401` → the API key was rejected. Do not retry; ask the user to download the skill again in Kkal Tracker (Settings → AI) and install it in place of this one.
- `429` → rate limited. Wait a few seconds and retry once.
- `5xx`, timeout or network error on `add` → the meal may have been saved anyway. Run `entries` and look for it before resending; a blind retry is how a meal gets logged twice.
