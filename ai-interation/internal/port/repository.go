package port

import "context"

type Repository interface {
	SaveMealAnalysis(ctx context.Context, imageName string, analysisJSON string) error

	GetUserProfile(ctx context.Context, userID string) (UserProfile, error)
	GetMealLogs(ctx context.Context, userID string, targetDate string) ([]MealLog, error)
	GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]ActivityLog, error)
	GetDietaryHistory(ctx context.Context, userID string, targetDate string) ([]DietaryHistory, error)
}

type UserProfile struct {
	UserID        string   `json:"user_id"`
	HeightCm      float64  `json:"height_cm"`
	WeightKg      float64  `json:"weight_kg"`
	ExerciseHabit string   `json:"exercise_habit"`
	HealthGoal    string   `json:"health_goal"`
	Allergies     []string `json:"allergies"`
	Preferences   []string `json:"preferences"`
}

type MealLog struct {
	Date     string  `json:"date"`
	MenuName string  `json:"menu_name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type ActivityLog struct {
	Date         string `json:"date"`
	ActivityName string `json:"activity_name"`
	DurationMin  int    `json:"duration_min"`
}

type DietaryHistory struct {
	FoodName  string   `json:"food_name"`
	MealType  string   `json:"meal_type"`
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
}