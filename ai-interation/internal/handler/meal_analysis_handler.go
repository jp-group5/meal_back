package handler

import (
	"net/http"

	"ai-interation/internal/usecase"

	"github.com/gin-gonic/gin"
)

type MealAnalysisHandler struct {
	usecase *usecase.MealAnalysisUsecase
}

func NewMealAnalysisHandler(usecase *usecase.MealAnalysisUsecase) *MealAnalysisHandler {
	return &MealAnalysisHandler{usecase: usecase}
}

func (h *MealAnalysisHandler) Analyze(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "image file is required in form field 'image'",
		})
		return
	}

	resp, err := h.usecase.Analyze(c.Request.Context(), fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
