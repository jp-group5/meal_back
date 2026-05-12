package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-interation/internal/dto"
	"ai-interation/internal/usecase"
)

type RecommendationHandler struct {
	usecase *usecase.RecommendationUsecase
}

func NewRecommendationHandler(u *usecase.RecommendationUsecase) *RecommendationHandler {
	return &RecommendationHandler{usecase: u}
}

func (h *RecommendationHandler) Generate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	fmt.Printf("[request body] %s\n", string(body))

	// ここで戻しておかないと ShouldBindJSON が読めなくなる
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var req dto.RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	res, err := h.usecase.Generate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}