# Meal App Backend Role Spec (Reusable)

## 1. Product Goal
- Users interact with a web calendar and record:
- Activities (`activity`)
- Meals (`meal`)
- The AI generates nutrition recommendations based on:
- User profile data
- Recent meal records
- This week's activity records

Backend responsibilities:
1. Manage user and business data in the database.
2. Expose Go APIs for frontend integration.
3. Build prompt-ready JSON from user/profile/meal/activity data.

## 2. Backend Scope
1. Auth and identity
- Register, login, refresh token, logout, get current user.
- JWT + persistent sessions (`user_sessions`).

2. User profile
- Height, weight, training experience, fitness goal, monthly budget.
- Preferences: `allergies`, `dietary_preferences`.

3. Meal records
- CRUD by date.
- Support AI acceptance metadata (`source`, `recommendationId`).

4. Activity records
- CRUD by date.
- Intensity (`low|medium|high`) for nutrition reasoning.

5. Recommendation + prompt assembly
- Input: date.
- Aggregate:
- User profile and preferences
- Recent 7-day meals
- Current week activities (Mon-Sun)
- Output:
- Frontend-ready recommendation object
- `prompt_json` for model input

## 3. API Contract (`/api/v1`)
- `PUT /users/me/preferences`
- `GET /meals?date=YYYY-MM-DD`
- `POST /meals`
- `PUT /meals/:id`
- `DELETE /meals/:id`
- `GET /activities?date=YYYY-MM-DD`
- `POST /activities`
- `PUT /activities/:id`
- `DELETE /activities/:id`
- `POST /recommendations` with body `{ "date": "YYYY-MM-DD" }`
- `GET /recommendations/prompt?date=YYYY-MM-DD` (prompt debug)

All endpoints above require `Authorization: Bearer <access_token>`.

## 4. Core Tables
- `users`
- `user_profiles` (includes JSONB `allergies`, `dietary_preferences`)
- `user_sessions`
- `meal_records`
- `activity_records`

## 5. Prompt JSON Shape (Example)
```json
{
  "metadata": {
    "generated_at": "2026-05-12T12:00:00+09:00",
    "target_date": "2026-05-12",
    "time_zone": "Asia/Tokyo"
  },
  "user": {
    "user_id": 1,
    "username": "alice",
    "height_cm": 168,
    "weight_kg": 56,
    "fitness_goal": "lose_weight",
    "training_experience": ["fitness", "yoga"],
    "monthly_food_budget": 2500,
    "allergies": ["peanut"],
    "dietary_preferences": ["high_protein"]
  },
  "context": {
    "recent_meals": [],
    "weekly_activities": [],
    "meal_stats": {
      "days_with_meals": 0,
      "total_meal_records": 0,
      "average_daily_calories": 0,
      "records_with_calories": 0,
      "calories_window_in_days": 7,
      "recent_meals_window_days": 7
    },
    "activity_intensity_stats": {
      "low": 0,
      "medium": 0,
      "high": 0,
      "unknown": 0
    }
  }
}
```

## 6. Runtime Prompt Example (Directly for AI)
Use this when calling the model with `prompt_json`.

```text
[System]
You are a nutrition recommendation assistant.
Use only the provided JSON context.
Do not provide medical diagnosis.
Strictly avoid listed allergies and respect dietary preferences.
Return valid JSON only.

Required output schema:
{
  "date": "YYYY-MM-DD",
  "choices": [
    {
      "title": "string",
      "reason": "string",
      "suggestedMeals": [
        {"type":"breakfast|lunch|dinner|snack","content":"string"}
      ]
    }
  ]
}

Rules:
- Return 2 to 4 choices.
- Each choice must include breakfast, lunch, and dinner.
- Keep reason concise and practical.

[User]
Here is the context JSON:
{{prompt_json}}
```

Example model output:

```json
{
  "date": "2026-05-12",
  "choices": [
    {
      "title": "Option A - Balanced Performance",
      "reason": "Balanced calories and protein for recovery with moderate carbs.",
      "suggestedMeals": [
        {"type": "breakfast", "content": "Greek yogurt oatmeal bowl + boiled egg + berries"},
        {"type": "lunch", "content": "Chicken and brown rice salad bowl + avocado"},
        {"type": "dinner", "content": "Steamed fish + broccoli + sweet potato"}
      ]
    },
    {
      "title": "Option B - Higher Protein",
      "reason": "Higher protein density to support muscle recovery and satiety.",
      "suggestedMeals": [
        {"type": "breakfast", "content": "Egg white omelet + cottage cheese + apple"},
        {"type": "lunch", "content": "Turkey quinoa bowl + mixed greens"},
        {"type": "dinner", "content": "Grilled salmon + asparagus + lentils"}
      ]
    }
  ]
}
```

## 7. Reusable AI Task Template
```text
You are my Go backend copilot for meal_app. Please do the following:
1) Keep API compatibility with frontend endpoints: /meals, /activities, /recommendations, /users/me/preferences.
2) Use PostgreSQL + GORM for user/profile/meal/activity/preference data.
3) In recommendation flow, aggregate profile + recent 7-day meals + current-week activities and output prompt_json.
4) Keep JWT auth and unified response format (code/message/data) for all new endpoints.
5) Return changed file list, key schema/field updates, and minimal integration test steps.
```

## 8. Minimal Integration Steps
1. Set env vars: `DB_DSN`, `JWT_SECRET`.
2. Start backend (default `:8080`).
3. Point frontend `VITE_API_BASE_URL` to `http://localhost:8080/api/v1`.
4. After login, call:
- `POST /recommendations` for recommendation payload + `prompt_json`
- `GET /recommendations/prompt?date=...` for raw prompt data only
