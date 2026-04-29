package router

import (
	"github.com/gin-gonic/gin"

	"ai-interation/internal/handler"
)

func RegisterRoutes(r *gin.Engine, mealHandler *handler.MealAnalysisHandler, recHandler *handler.RecommendationHandler) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/meal-analysis", mealHandler.Analyze)
		v1.POST("/recommendation", recHandler.Generate)
	}
}