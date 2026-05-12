package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ai-interation/internal/config"
	"ai-interation/internal/handler"
	"ai-interation/internal/infrastructure/mock"
	openaiinfra "ai-interation/internal/infrastructure/openai"
	"ai-interation/internal/router"
	"ai-interation/internal/usecase"
)

func main() {
	cfg := config.Load()

	aiClient, err := openaiinfra.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	repo := mock.NewRepositoryMock()

	mealUC := usecase.NewMealAnalysisUsecase(aiClient, repo)
	recUC := usecase.NewRecommendationUsecase(aiClient, repo)

	r := gin.Default()
	router.RegisterRoutes(
		r,
		handler.NewMealAnalysisHandler(mealUC),
		handler.NewRecommendationHandler(recUC),
	)

	log.Printf("server starting on %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}