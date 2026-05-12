package dto

import "ai-interation/internal/port"

type RecommendationRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	TargetDate string `json:"target_date" binding:"required"`
	Condition  string `json:"condition" binding:"required"`
}

type RecommendationResponse struct {
	Contents       []port.RecommendationItem `json:"contents"`
	TotalNutrition  port.Nutrition           `json:"total_nutrition"`
	Error          any                      `json:"error"`
}