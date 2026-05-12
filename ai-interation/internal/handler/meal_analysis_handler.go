package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-interation/internal/usecase"
)

type MealAnalysisHandler struct {
	usecase *usecase.MealAnalysisUsecase
}

func NewMealAnalysisHandler(u *usecase.MealAnalysisUsecase) *MealAnalysisHandler {
	return &MealAnalysisHandler{usecase: u}
}

func (h *MealAnalysisHandler) Analyze(c *gin.Context) {
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read image"})
		return
	}

	res, err := h.usecase.Analyze(c.Request.Context(), imageBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}