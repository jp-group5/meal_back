package mock

import (
	"context"

	"ai-interation/internal/port"
)

type RepositoryMock struct{}

func NewRepositoryMock() *RepositoryMock {
	return &RepositoryMock{}
}

func (r *RepositoryMock) SaveMealAnalysis(ctx context.Context, imageName string, analysisJSON string) error {
	return nil
}

func (r *RepositoryMock) GetUserProfile(ctx context.Context, userID string) (port.UserProfile, error) {
	return port.UserProfile{
		UserID:        userID,
		HeightCm:      170,
		WeightKg:      65,
		ExerciseHabit: "light",
	}, nil
}

func (r *RepositoryMock) GetMealLogs(ctx context.Context, userID string, targetDate string) ([]port.MealLog, error) {
	return []port.MealLog{
		{Date: targetDate, MenuName: "朝食: トースト", Calories: 320, Protein: 12, Fat: 10, Carbs: 45},
		{Date: targetDate, MenuName: "昼食: カレー", Calories: 720, Protein: 20, Fat: 24, Carbs: 95},
	}, nil
}

func (r *RepositoryMock) GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]port.ActivityLog, error) {
	return []port.ActivityLog{
		{Date: targetDate, ActivityName: "walking", DurationMin: 30},
	}, nil
}

var _ port.Repository = (*RepositoryMock)(nil)

func (r *RepositoryMock) GetDietaryHistory(ctx context.Context, userID string, targetDate string) ([]port.DietaryHistory, error) {
	return []port.DietaryHistory{
		{
			FoodName:  "Chicken Breast Salad",
			MealType:  "lunch",
			Timestamp: "2026-04-28T12:30:00Z",
			Tags:      []string{"high_protein", "clean_eating"},
		},
		{
			FoodName:  "Double Cheese Burger",
			MealType:  "dinner",
			Timestamp: "2026-04-27T19:00:00Z",
			Tags:      []string{"high_fat", "heavy"},
		},
		{
			FoodName:  "Yogurt and Banana",
			MealType:  "breakfast",
			Timestamp: "2026-04-26T08:00:00Z",
			Tags:      []string{"light", "easy"},
		},
	}, nil
}