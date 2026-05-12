package port

import "context"

type UserProfile struct {
	HeightCm           float64  `json:"height_cm"`
	WeightKg           float64  `json:"weight_kg"`
	FitnessGoal        string   `json:"fitness_goal"`
	ExerciseExperience []string  `json:"exercise_experience"`
	MonthlyFoodBudget  int       `json:"monthly_food_budget"`
	Allergies          []string  `json:"allergies"`
	DietaryPreferences []string  `json:"dietary_preferences"`
}

type MealLog struct {
	Date     string   `json:"date"`
	Content  []string  `json:"content"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Fat      float64   `json:"fat"`
	Carbs    float64   `json:"carbs"`
}

type ActivityLog struct {
	Date         string `json:"date"`
	ActivityName string `json:"activity_name"`
	DurationMin  int    `json:"duration_min"`
}

type AnalyzedMeal struct {
	Date     string   `json:"date"`
	Content  []string  `json:"content"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Fat      float64   `json:"fat"`
	Carbs    float64   `json:"carbs"`
}

type Nutrition struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type RecommendationItem struct {
	MenuName string  `json:"menu_name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	Reason   string  `json:"reason"`
}

type Repository interface {
	SaveMealAnalysis(ctx context.Context, meal AnalyzedMeal) error
	GetUserProfile(ctx context.Context, userID string) (UserProfile, error)
	GetMealLogs(ctx context.Context, userID string, targetDate string) ([]MealLog, error)
	GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]ActivityLog, error)
}