package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type MealAnalysisUsecase struct {
	aiClient   port.AIClient
	repository port.Repository
}

func NewMealAnalysisUsecase(aiClient port.AIClient, repository port.Repository) *MealAnalysisUsecase {
	return &MealAnalysisUsecase{aiClient: aiClient, repository: repository}
}

func (u *MealAnalysisUsecase) Analyze(ctx context.Context, imageBytes []byte) (*dto.MealAnalysisResponse, error) {
	raw, err := u.aiClient.AnalyzeMeal(ctx, imageBytes)
	if err != nil {
		return nil, err
	}

	var meal port.AnalyzedMeal
	if err := json.Unmarshal([]byte(raw), &meal); err != nil {
		return nil, fmt.Errorf("failed to parse meal analysis: %w; raw=%q", err, raw)
	}

	if err := u.repository.SaveMealAnalysis(ctx, meal); err != nil {
		return nil, err
	}

	return &dto.MealAnalysisResponse{
		Date:     meal.Date,
		Content:  meal.Content,
		Calories: meal.Calories,
		Protein:  meal.Protein,
		Fat:      meal.Fat,
		Carbs:    meal.Carbs,
		Error:    nil,
	}, nil
}