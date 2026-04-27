package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type RecommendationUsecase struct {
	aiClient port.AIClient
	repo     port.RecommendationRepository
}

func NewRecommendationUsecase(aiClient port.AIClient, repo port.RecommendationRepository) *RecommendationUsecase {
	return &RecommendationUsecase{
		aiClient: aiClient,
		repo:     repo,
	}
}

func (u *RecommendationUsecase) Generate(ctx context.Context, req dto.RecommendationRequest) (dto.RecommendationResponse, error) {
	inputs, err := u.repo.GetRecommendationInputs(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return dto.RecommendationResponse{}, err
	}

	prompt := buildRecommendationPrompt(req, inputs)

	rawText, err := u.aiClient.GenerateRecommendations(ctx, prompt)
	if err != nil {
		return dto.RecommendationResponse{}, err
	}

	var resp dto.RecommendationResponse
	if err := json.Unmarshal([]byte(rawText), &resp); err != nil {
		return dto.RecommendationResponse{}, err
	}

	return resp, nil
}

func buildRecommendationPrompt(req dto.RecommendationRequest, inputs port.RecommendationInputs) string {
	return fmt.Sprintf(
		"UserID: %s\nTargetDate: %s\nCondition: %s\nProfile: %+v\nActivities: %+v\nMealRecords: %+v",
		req.UserID,
		req.TargetDate,
		req.Condition,
		inputs.Profile,
		inputs.Activities,
		inputs.MealRecords,
	)
}
