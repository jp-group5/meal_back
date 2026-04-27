package port

import "context"

type AIClient interface {
	AnalyzeMeal(ctx context.Context, imageBytes []byte, filename string) (string, error)
	GenerateRecommendations(ctx context.Context, prompt string) (string, error)
}
