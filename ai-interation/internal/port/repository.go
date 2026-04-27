package port

import (
	"context"

	"ai-interation/internal/dto"
)

type MealAnalysisRecord struct {
	Filename string
	Response dto.MealAnalysisResponse
	RawText  string
}

type MealAnalysisRepository interface {
	SaveMealAnalysis(ctx context.Context, record MealAnalysisRecord) error
}

type UserProfile struct {
	UserID          string
	HeightCm        float64
	WeightKg        float64
	ExerciseHabit   string
	PreferenceNotes string
}

type ActivityLog struct {
	Date        string
	Description string
}

type MealRecord struct {
	Date   string
	Meal   string
	Notes  string
}

type RecommendationInputs struct {
	Profile     UserProfile
	Activities  []ActivityLog
	MealRecords []MealRecord
}

type RecommendationRepository interface {
	GetRecommendationInputs(ctx context.Context, userID string, targetDate string) (RecommendationInputs, error)
}
