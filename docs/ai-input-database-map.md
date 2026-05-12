# AI Input Database Map

| Category | Table | Field | Type | Required for AI | Description | Example |
|---|---|---|---|---|---|---|
| User identity | `users` | `id` | bigint/uint | Yes | Internal user key for joins. | `1` |
| User identity | `users` | `username` | varchar | Yes | Display/user identifier in prompt context. | `"alice"` |
| Profile | `user_profiles` | `user_id` | bigint/uint | Yes | FK to `users.id`. | `1` |
| Profile | `user_profiles` | `height_cm` | numeric(5,2) | Optional | Height for nutrition estimation. | `168.0` |
| Profile | `user_profiles` | `weight_kg` | numeric(5,2) | Optional | Weight for nutrition estimation. | `56.0` |
| Profile | `user_profiles` | `fitness_goal` | varchar(32) | Optional | Goal enum: `lose_weight`, `build_muscle`, `maintain_shape`. | `"build_muscle"` |
| Profile | `user_profiles` | `exercise_experience` | text | Optional | Raw training experience text. | `"fitness,yoga"` |
| Profile | `user_profiles` | `monthly_diet_budget` | bigint | Optional | Monthly food budget ceiling. | `2500` |
| Preferences | `user_profiles` | `allergies` | jsonb(array<string>) | Optional but recommended | Foods/ingredients to strictly avoid. | `["peanut","shellfish"]` |
| Preferences | `user_profiles` | `dietary_preferences` | jsonb(array<string>) | Optional | Preference hints (e.g. high protein). | `["high_protein"]` |
| Meal records | `meal_records` | `user_id` | bigint/uint | Yes | FK to user. | `1` |
| Meal records | `meal_records` | `date` | date | Yes | Meal date; used for recent 7-day window. | `"2026-05-12"` |
| Meal records | `meal_records` | `type` | varchar(16) | Yes | Enum: `breakfast`, `lunch`, `dinner`, `snack`. | `"dinner"` |
| Meal records | `meal_records` | `content` | text | Yes | Meal content text. | `"chicken rice bowl"` |
| Meal records | `meal_records` | `calories` | int | Optional | Calories for trend/stats. | `650` |
| Meal records | `meal_records` | `protein` | numeric(6,2) | Optional | Protein grams. | `42.0` |
| Meal records | `meal_records` | `carbs` | numeric(6,2) | Optional | Carbohydrate grams. | `58.0` |
| Meal records | `meal_records` | `fat` | numeric(6,2) | Optional | Fat grams. | `19.0` |
| Meal records | `meal_records` | `source` | varchar(64) | Optional | Record source (`manual`, `ai-recommendation`, etc.). | `"manual"` |
| Meal records | `meal_records` | `recommendation_id` | varchar(128) | Optional | Traceability to AI recommendation. | `"rec-1770000000"` |
| Activities | `activity_records` | `user_id` | bigint/uint | Yes | FK to user. | `1` |
| Activities | `activity_records` | `date` | date | Yes | Activity date; used for current-week window. | `"2026-05-12"` |
| Activities | `activity_records` | `title` | varchar(128) | Yes | Activity label. | `"gym"` |
| Activities | `activity_records` | `start_time` | varchar(8) | Optional | Start time `HH:MM`. | `"18:00"` |
| Activities | `activity_records` | `end_time` | varchar(8) | Optional | End time `HH:MM`. | `"19:30"` |
| Activities | `activity_records` | `intensity` | varchar(16) | Optional but recommended | Enum: `low`, `medium`, `high`. | `"high"` |

## Notes for teammate

| Item | Rule |
|---|---|
| Recommended data scope | Use profile + recent 7-day meals + current-week activities. |
| Must-have joins | `users.id = user_profiles.user_id = meal_records.user_id = activity_records.user_id`. |
| Session table | `user_sessions` is auth/session data; normally not needed for nutrition recommendation prompts. |
