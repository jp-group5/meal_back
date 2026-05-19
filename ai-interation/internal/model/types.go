package model

type RecommendationRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	TargetDate string `json:"target_date" binding:"required"`
	Condition  string `json:"condition" binding:"required"`
}

type UserProfile struct {
	HeightCm           float64  `json:"height_cm"`
	WeightKg           float64  `json:"weight_kg"`
	FitnessGoal        string   `json:"fitness_goal"`
	ExerciseExperience []string `json:"exercise_experience"`
	MonthlyFoodBudget  int      `json:"monthly_food_budget"`
	Allergies          []string `json:"allergies"`
	DietaryPreferences []string `json:"dietary_preferences"`
}

type MealLog struct {
	Date     string   `json:"date"`
	Content  []string `json:"content"`
	Calories float64  `json:"calories"`
	Protein  float64  `json:"protein"`
	Fat      float64  `json:"fat"`
	Carbs    float64  `json:"carbs"`
}

type ActivityLog struct {
	Date         string `json:"date"`
	ActivityName string `json:"activity_name"`
	DurationMin  int    `json:"duration_min"`
}

type Nutrition struct {
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type MealAnalysisResponse struct {
	Contents       []string  `json:"contents"`
	TotalNutrition Nutrition `json:"total_nutrition"`
	Error          any       `json:"error"`
}

type AnalyzedMeal struct {
	Date     string   `json:"date"`
	Content  []string `json:"content"`
	Calories float64  `json:"calories"`
	Protein  float64  `json:"protein"`
	Fat      float64  `json:"fat"`
	Carbs    float64  `json:"carbs"`
}

type Recommendation struct {
	MenuName string  `json:"menu_name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
	Reason   string  `json:"reason"`
}

type RecommendationResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
	Error           any              `json:"error"`
}