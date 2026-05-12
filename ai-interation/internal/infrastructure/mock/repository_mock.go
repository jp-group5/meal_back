package mock

import (
	"context"
	"time"

	"ai-interation/internal/port"
)

type RepositoryMock struct{}

func NewRepositoryMock() *RepositoryMock { return &RepositoryMock{} }

func (r *RepositoryMock) SaveMealAnalysis(ctx context.Context, meal port.AnalyzedMeal) error {
	return nil
}

func (r *RepositoryMock) GetUserProfile(ctx context.Context, userID string) (port.UserProfile, error) {
	return port.UserProfile{
		HeightCm:           170,
		WeightKg:           65,
		FitnessGoal:        "muscle_gain",
		ExerciseExperience: []string{"beginner"},
		MonthlyFoodBudget:  40000,
		Allergies:          []string{"peanuts", "shellfish"},
		DietaryPreferences: []string{"low_carb", "no_cilantro"},
	}, nil
}

func (r *RepositoryMock) GetMealLogs(ctx context.Context, userID string, targetDate string) ([]port.MealLog, error) {
	d, _ := time.Parse("2006-01-02", targetDate)
	d = d.UTC()

	return []port.MealLog{
		{
			Date:     d.Format("2006-01-02"),
			Content:  []string{"Chicken Breast Salad"},
			Calories: 520,
			Protein:  42,
			Fat:      18,
			Carbs:    28,
		},
		{
			Date:     d.AddDate(0, 0, -1).Format("2006-01-02"),
			Content:  []string{"Double Cheese Burger"},
			Calories: 980,
			Protein:  34,
			Fat:      58,
			Carbs:    62,
		},
		{
			Date:     d.AddDate(0, 0, -2).Format("2006-01-02"),
			Content:  []string{"Yogurt and Banana"},
			Calories: 260,
			Protein:  12,
			Fat:      4,
			Carbs:    38,
		},
	}, nil
}

func (r *RepositoryMock) GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]port.ActivityLog, error) {
	return []port.ActivityLog{
		{Date: targetDate, ActivityName: "Morning Jogging", DurationMin: 45},
		{Date: targetDate, ActivityName: "Evening Coding Session", DurationMin: 120},
	}, nil
}

var _ port.Repository = (*RepositoryMock)(nil)