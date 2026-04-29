package dto

type RecommendationRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	TargetDate string `json:"target_date" binding:"required"`
	Condition  string `json:"condition" binding:"required"`
}

type RecommendationItem struct {
	MenuName string `json:"menu_name"`
	Reason   string `json:"reason"`
}

type RecommendationResponse struct {
	Recommendations []RecommendationItem `json:"recommendations"`
	Error           any                  `json:"error"`
}