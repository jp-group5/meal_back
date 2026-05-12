package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type RecommendationUsecase struct {
	aiClient   port.AIClient
	repository port.Repository
}

func NewRecommendationUsecase(aiClient port.AIClient, repository port.Repository) *RecommendationUsecase {
	return &RecommendationUsecase{aiClient: aiClient, repository: repository}
}

type dietaryHistoryForAI struct {
	Date     string   `json:"date"`
	DaysAgo  int      `json:"days_ago"`
	Content  []string  `json:"content"`
	Calories float64  `json:"calories"`
	Protein  float64  `json:"protein"`
	Fat      float64  `json:"fat"`
	Carbs    float64  `json:"carbs"`
}

type recommendationAIResponse struct {
	Recommendations []port.RecommendationItem `json:"recommendations"`
	Error           any                       `json:"error"`
}

func (u *RecommendationUsecase) Generate(ctx context.Context, req dto.RecommendationRequest) (*dto.RecommendationResponse, error) {
	profile, err := u.repository.GetUserProfile(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	mealLogs, err := u.repository.GetMealLogs(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return nil, err
	}

	activityLogs, err := u.repository.GetActivityLogs(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return nil, err
	}

	targetDay, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target_date: %w", err)
	}
	targetDay = time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), 0, 0, 0, 0, time.UTC)

	payload := map[string]any{
		"condition": req.Condition,
		"user_meta": profile,
		"today_context": map[string]any{
			"current_time": targetDay.Format(time.RFC3339),
			"activities":   activityLogs,
		},
		"dietary_history": convertMealLogsToHistory(targetDay, mealLogs),
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`以下のデータをもとに、おすすめ献立を3件、JSONのみで返してください。
- allergies を必ず避ける
- dietary_preferences を考慮する
- fitness_goal に合うものを優先する
- dietary_history の days_ago は 0 が当日、1 が前日、2 が2日前

データ:
%s`, string(b))

	raw, err := u.aiClient.GenerateRecommendation(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var aiRes recommendationAIResponse
	if err := json.Unmarshal([]byte(raw), &aiRes); err != nil {
		return nil, fmt.Errorf("failed to parse recommendation: %w; raw=%q", err, raw)
	}

	var total port.Nutrition
	for _, r := range aiRes.Recommendations {
		total.Calories += r.Calories
		total.Protein += r.Protein
		total.Fat += r.Fat
		total.Carbs += r.Carbs
	}

	return &dto.RecommendationResponse{
		Contents:      aiRes.Recommendations,
		TotalNutrition: total,
		Error:         nil,
	}, nil
}

func convertMealLogsToHistory(targetDay time.Time, logs []port.MealLog) []dietaryHistoryForAI {
	out := make([]dietaryHistoryForAI, 0, len(logs))

	for _, log := range logs {
		d, err := time.Parse("2006-01-02", log.Date)
		if err != nil {
			continue
		}
		d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)

		daysAgo := int(targetDay.Sub(d).Hours() / 24)
		if daysAgo < 0 {
			daysAgo = 0
		}

		out = append(out, dietaryHistoryForAI{
			Date:     log.Date,
			DaysAgo:  daysAgo,
			Content:  log.Content,
			Calories: log.Calories,
			Protein:  log.Protein,
			Fat:      log.Fat,
			Carbs:    log.Carbs,
		})
	}

	return out
}