package usecase

import (
	"context"
	"encoding/json"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type MealAnalysisUsecase struct {
	aiClient   port.AIClient
	repository port.Repository
}

func NewMealAnalysisUsecase(aiClient port.AIClient, repository port.Repository) *MealAnalysisUsecase {
	return &MealAnalysisUsecase{
		aiClient:   aiClient,
		repository: repository,
	}
}

func (u *MealAnalysisUsecase) Analyze(ctx context.Context, imageBytes []byte, imageName string) (*dto.MealAnalysisResponse, error) {
	raw, err := u.aiClient.AnalyzeMeal(ctx, imageBytes)
	if err != nil {
		return nil, err
	}

	var res dto.MealAnalysisResponse
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, err
	}

	res.Error = nil

	if u.repository != nil {
		_ = u.repository.SaveMealAnalysis(ctx, imageName, raw)
	}

	return &res, nil
}