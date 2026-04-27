package router

import (
	"ai-interation/internal/handler"
	"ai-interation/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(mealAnalysisUsecase *usecase.MealAnalysisUsecase, recommendationUsecase *usecase.RecommendationUsecase) *gin.Engine {
	r := gin.Default()

	mealAnalysisHandler := handler.NewMealAnalysisHandler(mealAnalysisUsecase)
	recommendationHandler := handler.NewRecommendationHandler(recommendationUsecase)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/meal-analysis", mealAnalysisHandler.Analyze)
		v1.POST("/recommendation", recommendationHandler.Generate)
	}

	return r
}
