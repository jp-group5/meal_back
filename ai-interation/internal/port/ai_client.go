package port

import "context"

type AIClient interface {
	AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error)
	GenerateRecommendation(ctx context.Context, prompt string) (string, error)
}