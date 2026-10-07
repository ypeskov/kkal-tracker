# Features

- **User Authentication**: JWT-based login/register/activate with bcrypt password hashing and email activation
- **Persistent Sessions**: JWT tokens stored in sessionStorage, survive page reloads
- **Calorie Tracking**: food entries with name, weight, kcal/100g, auto-calculated totals
- **Dashboard**: today's entries with the total calorie count, weight goal card
- **Weight Management**: weight history over time with charts
- **Weight Goals**: target weight with optional target date and a start date, progress with visual indicators.
  Editing a goal keeps its start date and initial weight unless the start date is changed; the initial weight is the
  latest weight on or before the start date. The report chart shows a trend line: a least-squares fit through the
  initial weight and every weigh-in after the start day (`trend` in the goal progress), extended to the target date
- **Health Metrics**: BMI, BMR, TDEE based on the user profile (age, gender, activity level)
- **Ingredient Database**: global ingredients with multilingual names and nutritional data
- **User Profiles**: personal settings, preferences, gender, activity level
- **Reports & Analytics**: calorie and weight trends
- **AI Insights**: AI-powered nutrition and weight analysis with personalized recommendations
  - OpenAI integration (GPT-4o-mini by default), provider selection in the UI (extensible to more providers)
  - Analysis periods of 7, 14, 30, 90 days, optional specific question
  - Answers in the user's language
  - Rate limited: 2 requests per minute
- **Data Export**: weight and food data as Excel files (download or email delivery)
- **API Key Data Access**: programmatic access to weight/food data
  - Key management UI in Settings (create, revoke, delete)
  - SHA-256 hashed storage, the full key is shown only once at creation
  - Time-limited (N days) or permanent keys
  - `X-API-Key` header authentication for the `/api/v1` endpoints
  - A key gives full access to the user's data, both read and write (there are no scopes)
- **Food Logging API for AI agents**: an agent (e.g. voice dictation through a skill from `skills/`) stores meals via `/api/v1`
  - The agent reads the whole ingredient list (`GET /api/v1/ingredients`, with usage statistics) and matches foods itself
  - `POST /api/v1/food` takes a meal as a list of items: `ingredient_id` for an existing ingredient or `new_ingredient`
    for a new one. The server never matches or creates ingredients by name on its own
  - A new ingredient whose name equals an existing one (ignoring case, spacing and `ё`/`е`) is rejected with 409 and
    the existing ingredient
  - `one_off: true` writes a new food to the diary without adding it to the ingredient list (restaurant dishes etc.)
  - A meal is atomic (all items or none); calories are calculated on the server
  - `PUT /api/v1/food/:id` corrects an entry (weight, another ingredient, calorie value, time),
    `DELETE /api/v1/food/:id` removes it
- **Email Service**: activation emails and export delivery
- **Internationalization**: en_US, uk_UA, ru_UA, bg_BG with a language switcher, frontend and backend
- **Integrated Calculator**: built into every numeric input (weight, calories, macros), see the `frontend-style` skill
- **Database Backup**: nightly backups to Google Drive, see [`infrastructure.md`](infrastructure.md)
