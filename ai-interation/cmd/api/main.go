package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ai-interation/internal/handler"
	"ai-interation/internal/infrastructure/mock"
	openaiinfra "ai-interation/internal/infrastructure/openai"
	"ai-interation/internal/router"
	"ai-interation/internal/usecase"
)

func main() {
	aiClient, err := openaiinfra.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	repo := mock.NewRepositoryMock()

	mealUC := usecase.NewMealAnalysisUsecase(aiClient, repo)
	recUC := usecase.NewRecommendationUsecase(aiClient, repo)

	mealHandler := handler.NewMealAnalysisHandler(mealUC)
	recHandler := handler.NewRecommendationHandler(recUC)

	r := gin.Default()
	router.RegisterRoutes(r, mealHandler, recHandler)

	log.Println("server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}