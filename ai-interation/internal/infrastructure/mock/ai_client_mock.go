package mock

import (
	"context"
	"encoding/json"

	"ai-interation/internal/dto"
)

type AIClientMock struct{}

func NewAIClientMock() *AIClientMock {
	return &AIClientMock{}
}

func (m *AIClientMock) AnalyzeMeal(ctx context.Context, imageBytes []byte, filename string) (string, error) {
	resp := dto.MealAnalysisResponse{
		Items: []dto.MealItem{
			{Name: "焼き鮭"},
			{Name: "みそ汁"},
			{Name: "漬物"},
			{Name: "白米"},
		},
		TotalNutrition: dto.TotalNutrition{
			Calories:        490,
			Protein:         27.4,
			Fat:             10.8,
			Carbohydrates:   62.0,
			VegetablesAmount: 40.0,
		},
		Error: nil,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m *AIClientMock) GenerateRecommendations(ctx context.Context, prompt string) (string, error) {
	resp := dto.RecommendationResponse{
		Recommendations: []dto.RecommendationItem{
			{
				MenuName: "鶏むね肉のサラダ定食",
				Reason:   "たんぱく質を確保しつつ、脂質を抑えやすいため",
			},
			{
				MenuName: "鮭と野菜の蒸し料理",
				Reason:   "野菜量を増やしやすく、栄養バランスを整えやすいため",
			},
			{
				MenuName: "豆腐と野菜のうどん",
				Reason:   "消化しやすく、簡単に用意できるため",
			},
		},
		Error: nil,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
