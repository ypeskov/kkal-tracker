---
name: kkal-tracker-food-log
description: Record what the user ate or drank into their Kkal Tracker food diary (kcal.peskov.info) through its API. Use this skill whenever the user tells you what they ate, dictates a meal or a snack, lists foods with weights, or asks to log, add, write down or save food or calories — including casual or voice-dictated phrasings like "я съел 200 грамм курицы и помидор", "запиши обед", "на завтрак было…", "добавь в дневник", "log my lunch", "I had two eggs and toast". Also use it to correct a recorded entry (weight, food, calories, time), to undo or delete one, and to check what is already logged today. The user does not need to mention Kkal Tracker by name — any "I ate X" or "запиши еду" request routes here. This is the default for food logging — the "-dev" copy is only for a message that itself names the dev or test tracker, and an earlier dev request in the conversation does not carry over. Edit or delete an entry with the copy that created it.
---

# Kkal Tracker: food diary logging

This skill writes meals into the user's Kkal Tracker diary. The user usually dictates by voice: one phrase may contain several foods, the names are approximate, and weights may be missing. Your job is to turn that phrase into exact diary entries without making the user repeat themselves.

The server is deliberately strict: it never guesses. Every food you send is either an ingredient the user already has (referenced by `ingredient_id`) or an explicitly described new food. Choosing between the two is your part of the work.

## Production and dev: where a request belongs

Two copies of this skill can be installed: the regular one writes to the real diary, the `-dev` one to a test server. A mix-up either puts test data into the real diary or loses a real meal in the test database, so the routing is strict.

- **Logging food** — every message is routed on its own. It goes to dev only when that message itself names the dev or test tracker ("запиши на дев", "в тестовый трекер", "log it to dev"). A dev request earlier in the conversation does not carry over: a following "добавь ещё кофе" without the word goes to the real diary, even in the middle of a testing session.
- **Editing and deleting** — these follow the entry, not the wording. Use the copy that created the entry, however the request is phrased; replies of the dev copy start with `[DEV]`, so the conversation shows where each entry went. Entry IDs are separate on each server and the same number can belong to a different, real entry on the other one, so an edit or delete sent to the wrong server silently damages an unrelated entry.
- If you cannot tell where an entry was created, find it with `entries` before changing anything, and ask the user when it is still unclear.

## API basics

- Base URL: `https://kcal.peskov.info`
- Auth: `X-API-Key` header. The client takes the key from the `KKAL_API_KEY` environment variable, from a file named `api_key` in the skill folder (next to this file), or from the constant at the top of `scripts/kkal_client.py`. If the script says the key is not configured, ask the user to create one in Kkal Tracker (Settings → API Keys) and save it to the `api_key` file.
- Rate limit: a burst of 20 requests, then 1 request per second. A normal logging takes two calls.

Use the bundled client, it handles the key, headers and errors:

```bash
python scripts/kkal_client.py ingredients            # the user's ingredient list
python scripts/kkal_client.py add --json '<meal>'    # store a meal (or pipe the JSON to stdin)
python scripts/kkal_client.py entries                # today's diary entries with IDs
python scripts/kkal_client.py entries --from 2026-03-01 --to 2026-03-07
python scripts/kkal_client.py edit 12345 --weight 150   # correct an entry
python scripts/kkal_client.py delete 12345           # delete an entry
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

If the user states a calorie value for a known food that differs from the list ("этот хлеб 260 ккал"), keep the `ingredient_id` and add `kcal_per_100g` to the item: it overrides the value for this entry only.

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
    {"ingredient_id": 94, "weight": 60, "kcal_per_100g": 260},
    {"new_ingredient": {"name": "Сырники", "kcal_per_100g": 220, "fats": 9, "carbs": 20, "proteins": 14}, "weight": 150},
    {"new_ingredient": {"name": "Паста карбонара", "kcal_per_100g": 180, "fats": 9, "carbs": 16, "proteins": 8}, "weight": 300, "one_off": true}
  ]
}
```

- Send all foods of the meal together. The request is atomic: if one item is rejected, nothing is saved, so you can fix the item and resend the whole meal without creating duplicates.
- `meal_datetime` is optional. Omit it when the user is logging what they just ate. Set it (with the user's UTC offset) when they name a time: "на завтрак в 8", "вчера вечером".
- The server calculates the calories itself from weight and calories per 100 g — do not send totals.
- Limits: weight up to 10000 g, `kcal_per_100g` 0–900, macros 0–100 g per 100 g, up to 50 items.

The response lists the stored entries (with `id`, the resolved `food` name and `calories`), `total_calories` of this meal and `day_total_calories` for the diary day.

### 6. Reply to the user

The user is often on voice and wants a quick confirmation, not a report. Reply in the user's language, in a few lines:

- what was recorded: food, grams, calories for each item;
- the meal total and the day total;
- anything they should know to catch a mistake: weights you assumed, which ingredient you picked when there was a choice, new ingredients created (with the values you estimated), one-off items.

Example:

```
Записал:
• Свиная шея (вес сырой) — 200 г, 534 ккал
• Помидор — 120 г, 24 ккал (вес одного помидора принял за 120 г)
• Сырники — 150 г, 330 ккал — новый продукт в справочнике: 220 ккал/100 г, Б 14 / Ж 9 / У 20 (оценка)

Приём пищи: 888 ккал. За день: 1 640 ккал.
```

## Fixing mistakes

Use the entry `id` from the `add` response. If you no longer have it (another conversation, an entry made on the website), find it with `entries`; when several entries could be the one the user means, ask rather than guess.

- **Correct an entry** ("исправь вес на 150", "там было не 200, а 120 грамм", "это была не шея, а грудка", "это было в обед, а не утром") → `edit`. It changes only what you pass and keeps the rest, including the meal time; the server recalculates the calories and returns the new day total.

  ```bash
  python scripts/kkal_client.py edit 12345 --weight 150
  python scripts/kkal_client.py edit 12345 --ingredient-id 76          # another food from the ingredient list
  python scripts/kkal_client.py edit 12345 --kcal 260                  # calories per 100 g for this entry only
  python scripts/kkal_client.py edit 12345 --datetime 2026-10-02T13:30:00+03:00
  ```

  The options can be combined. `--ingredient-id` takes the name, calories and nutrients from that ingredient and keeps the weight.
- **The right food is not in the ingredient list** → `edit` cannot create ingredients. Delete the entry and `add` it again as a new food, passing the original `meal_datetime` so the entry does not move to the current time.
- **Undo** ("отмени", "удали последнее", "это было не то") → `delete` each entry `id`.
- Ingredients created by mistake cannot be removed through this API; tell the user to delete them on the Food List page of the website.

After a correction, tell the user what the entry looks like now and the new day total.

## Error handling

- `409` → a new ingredient has the same name as an existing one (case, spacing and "ё/е" are ignored). The response contains `existing_ingredient` and `item_index`: replace that item's `new_ingredient` with the `ingredient_id` and resend the whole meal. Nothing was saved.
- `400` → the message names the item (`items[2]: …`) and the problem. Fix it and resend the whole meal. Nothing was saved.
- `404` on `edit` or `delete` → there is no such entry (already deleted, or a wrong ID). Check with `entries`.
- `401` → the API key was rejected. Tell the user, do not retry.
- `429` → rate limited. Wait a few seconds and retry once.
- `5xx`, timeout or network error on `add` → the meal may have been saved anyway. Run `entries` and look for it before resending; a blind retry is how a meal gets logged twice.
