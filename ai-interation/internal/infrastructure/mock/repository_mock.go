package mock

import (
	"context"
	"fmt"
	"sync"

	"ai-interation/internal/port"
)

type MealAnalysisRepositoryMock struct {
	mu      sync.Mutex
	records []port.MealAnalysisRecord
}

func NewMealAnalysisRepositoryMock() *MealAnalysisRepositoryMock {
	return &MealAnalysisRepositoryMock{
		records: make([]port.MealAnalysisRecord, 0),
	}
}

func (r *MealAnalysisRepositoryMock) SaveMealAnalysis(ctx context.Context, record port.MealAnalysisRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.records = append(r.records, record)
	fmt.Printf("[mock repository] saved meal analysis: filename=%s items=%d\n", record.Filename, len(record.Response.Items))
	return nil
}

type RecommendationRepositoryMock struct{}

func NewRecommendationRepositoryMock() *RecommendationRepositoryMock {
	return &RecommendationRepositoryMock{}
}

func (r *RecommendationRepositoryMock) GetRecommendationInputs(ctx context.Context, userID string, targetDate string) (port.RecommendationInputs, error) {
	return port.RecommendationInputs{
		Profile: port.UserProfile{
			UserID:          userID,
			HeightCm:        170,
			WeightKg:        65,
			ExerciseHabit:   "週2回の軽い運動",
			PreferenceNotes: "和食寄り、魚を好む",
		},
		Activities: []port.ActivityLog{
			{
				Date:        targetDate,
				Description: "通勤で徒歩移動が多い",
			},
			{
				Date:        targetDate,
				Description: "夕方に軽い散歩を実施",
			},
		},
		MealRecords: []port.MealRecord{
			{
				Date:  targetDate,
				Meal:  "朝食: トースト、卵、ヨーグルト",
				Notes: "炭水化物やや多め",
			},
			{
				Date:  targetDate,
				Meal:  "昼食: ラーメン",
				Notes: "脂質と塩分がやや多め",
			},
		},
	}, nil
}
