package httpapi

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/meal-analysis", h.AnalyzeMeal)
		v1.POST("/recommendation", h.GenerateRecommendation)
	}
}