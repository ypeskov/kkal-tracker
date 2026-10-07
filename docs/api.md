# API Endpoints

All routes are under `/api`.

- `GET /api/languages` - supported languages (public, no auth)
- `/api/auth/*` - authentication (rate limited: 5 req/sec per IP)
  - `POST /api/auth/login` - login
  - `POST /api/auth/register` - registration (additionally: 3 attempts, then 1 per 20 min per IP)
  - `GET /api/auth/activate/:token` - activate the account from the email link
  - `GET /api/auth/me` - current user (requires auth)
- `/api/calories/*` - calorie entry CRUD (GET, POST, PUT /:id, DELETE /:id)
- `/api/ingredients/*` - ingredient management (GET, GET /:id, POST, PUT /:id, DELETE /:id)
- `/api/weight/*` - weight history (GET, POST, PUT /:id, DELETE /:id)
- `/api/profile` - user profile (GET, PUT)
- `/api/profile/goal` - weight goal (GET progress, PUT set, DELETE clear)
- `/api/metrics` - health metrics (GET)
- `/api/reports/data` - aggregated report data (GET, `from`/`to` query params)
- `/api/ai/*` - AI analysis (rate limited: 2 req/min)
  - `GET /api/ai/status` - AI service status
  - `POST /api/ai/analyze` - run an analysis
- `/api/export` - data export (POST, weight/food as Excel or email)
- `/api/api-keys/*` - API key management (JWT auth)
  - `POST /api/api-keys` - create a key (returns the full key once)
  - `GET /api/api-keys` - list the user's keys (prefixes only)
  - `POST /api/api-keys/:id/revoke` - revoke a key
  - `DELETE /api/api-keys/:id` - delete a key
- `/api/v1/*` - external API (`X-API-Key` header, rate limited per IP: burst of 20, then 1 req/sec)
  - `GET /api/v1/data?type=weight|food|both&from=YYYY-MM-DD&to=YYYY-MM-DD`
  - `GET /api/v1/ingredients` - all user ingredients with `times_used` and `last_used`
  - `POST /api/v1/food` - store a meal: `{meal_datetime?, items: [{ingredient_id, weight, kcal_per_100g?} | {new_ingredient: {name, kcal_per_100g, fats?, carbs?, proteins?}, weight, one_off?}]}`
  - `PUT /api/v1/food/:id` - change a food entry: `{weight?, ingredient_id?, kcal_per_100g?, meal_datetime?}`;
    fields that are not sent keep their values, calories are recalculated
  - `DELETE /api/v1/food/:id` - delete a food entry
