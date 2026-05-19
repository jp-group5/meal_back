package repository

import (
	"context"
	"time"

	"ai-interation/internal/model"
)

type MockRepository struct{}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (r *MockRepository) SaveAnalyzedMeal(ctx context.Context, meal model.AnalyzedMeal) error {
	return nil
}

func (r *MockRepository) GetUserProfile(ctx context.Context, userID string) (model.UserProfile, error) {
	return model.UserProfile{
		HeightCm:           170,
		WeightKg:           65,
		FitnessGoal:        "muscle_gain",
		ExerciseExperience: []string{"beginner"},
		MonthlyFoodBudget:  40000,
		Allergies:          []string{"peanuts", "shellfish"},
		DietaryPreferences: []string{"low_carb", "no_cilantro"},
	}, nil
}

func (r *MockRepository) GetMealLogs(ctx context.Context, userID string, targetDate string) ([]model.MealLog, error) {
	target, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		target = time.Now()
	}

	return []model.MealLog{
		{
			Date:     target.Format("2006-01-02"),
			Content:  []string{"Chicken Breast Salad"},
			Calories: 520,
			Protein:  42,
			Fat:      18,
			Carbs:    28,
		},
		{
			Date:     target.AddDate(0, 0, -1).Format("2006-01-02"),
			Content:  []string{"Double Cheese Burger"},
			Calories: 980,
			Protein:  34,
			Fat:      58,
			Carbs:    62,
		},
		{
			Date:     target.AddDate(0, 0, -2).Format("2006-01-02"),
			Content:  []string{"Yogurt", "Banana"},
			Calories: 260,
			Protein:  12,
			Fat:      4,
			Carbs:    38,
		},
	}, nil
}

func (r *MockRepository) GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]model.ActivityLog, error) {
	return []model.ActivityLog{
		{
			Date:         targetDate,
			ActivityName: "Morning Jogging",
			DurationMin:  45,
		},
		{
			Date:         targetDate,
			ActivityName: "Coding Session",
			DurationMin:  120,
		},
	}, nil
}