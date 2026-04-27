package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type MealAnalysisUsecase struct {
	aiClient port.AIClient
	repo     port.MealAnalysisRepository
}

func NewMealAnalysisUsecase(aiClient port.AIClient, repo port.MealAnalysisRepository) *MealAnalysisUsecase {
	return &MealAnalysisUsecase{
		aiClient: aiClient,
		repo:     repo,
	}
}

func (u *MealAnalysisUsecase) Analyze(ctx context.Context, fileHeader *multipart.FileHeader) (dto.MealAnalysisResponse, error) {
	if fileHeader == nil {
		return dto.MealAnalysisResponse{}, errors.New("image file is required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return dto.MealAnalysisResponse{}, err
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		return dto.MealAnalysisResponse{}, err
	}

	rawText, err := u.aiClient.AnalyzeMeal(ctx, imageBytes, fileHeader.Filename)
	if err != nil {
		return dto.MealAnalysisResponse{}, err
	}

	var resp dto.MealAnalysisResponse
	if err := json.Unmarshal([]byte(rawText), &resp); err != nil {
		return dto.MealAnalysisResponse{}, err
	}

	if err := u.repo.SaveMealAnalysis(ctx, port.MealAnalysisRecord{
		Filename: fileHeader.Filename,
		Response: resp,
		RawText:  rawText,
	}); err != nil {
		return dto.MealAnalysisResponse{}, err
	}

	return resp, nil
}
