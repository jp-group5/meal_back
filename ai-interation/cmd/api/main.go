package main

import (
	"log"

	"ai-interation/internal/config"
	"ai-interation/internal/infrastructure/mock"
	"ai-interation/internal/router"
	"ai-interation/internal/usecase"
)

func main() {
	cfg := config.Load()

	mealAnalysisRepo := mock.NewMealAnalysisRepositoryMock()
	recommendationRepo := mock.NewRecommendationRepositoryMock()
	aiClient := mock.NewAIClientMock()

	mealAnalysisUsecase := usecase.NewMealAnalysisUsecase(aiClient, mealAnalysisRepo)
	recommendationUsecase := usecase.NewRecommendationUsecase(aiClient, recommendationRepo)

	r := router.NewRouter(mealAnalysisUsecase, recommendationUsecase)

	log.Printf("server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
