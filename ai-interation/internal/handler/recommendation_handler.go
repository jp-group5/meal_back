package handler

import (
	"net/http"

	"ai-interation/internal/dto"
	"ai-interation/internal/usecase"

	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	usecase *usecase.RecommendationUsecase
}

func NewRecommendationHandler(usecase *usecase.RecommendationUsecase) *RecommendationHandler {
	return &RecommendationHandler{usecase: usecase}
}

func (h *RecommendationHandler) Generate(c *gin.Context) {
	var req dto.RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resp, err := h.usecase.Generate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
