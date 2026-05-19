package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-interation/internal/model"
)

type AIClient interface {
	AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error)
	GenerateRecommendation(ctx context.Context, prompt string) (string, error)
}

type Repository interface {
	SaveAnalyzedMeal(ctx context.Context, meal model.AnalyzedMeal) error
	GetUserProfile(ctx context.Context, userID string) (model.UserProfile, error)
	GetMealLogs(ctx context.Context, userID string, targetDate string) ([]model.MealLog, error)
	GetActivityLogs(ctx context.Context, userID string, targetDate string) ([]model.ActivityLog, error)
}

type Service struct {
	ai   AIClient
	repo Repository
}

func NewService(ai AIClient, repo Repository) *Service {
	return &Service{ai: ai, repo: repo}
}

func (s *Service) AnalyzeMeal(ctx context.Context, imageBytes []byte) (*model.MealAnalysisResponse, error) {
	raw, err := s.ai.AnalyzeMeal(ctx, imageBytes)
	if err != nil {
		return nil, err
	}

	var res model.MealAnalysisResponse
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, fmt.Errorf("failed to parse meal analysis response: %w; raw=%q", err, raw)
	}

	res.Error = nil

	meal := model.AnalyzedMeal{
		Date:     time.Now().Format("2006-01-02"),
		Content:  res.Contents,
		Calories: res.TotalNutrition.Calories,
		Protein:  res.TotalNutrition.Protein,
		Fat:      res.TotalNutrition.Fat,
		Carbs:    res.TotalNutrition.Carbs,
	}

	if err := s.repo.SaveAnalyzedMeal(ctx, meal); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *Service) GenerateRecommendation(ctx context.Context, req model.RecommendationRequest) (*model.RecommendationResponse, error) {
	profile, err := s.repo.GetUserProfile(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	mealLogs, err := s.repo.GetMealLogs(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return nil, err
	}

	activityLogs, err := s.repo.GetActivityLogs(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return nil, err
	}

	payload, err := buildRecommendationPayload(req, profile, mealLogs, activityLogs)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`以下のデータをもとに、おすすめ献立を3件提案してください。

条件:
- allergies に含まれる食材は絶対に使わない
- dietary_preferences を考慮する
- fitness_goal に合う献立を優先する
- monthly_food_budget を考慮し、現実的な献立にする
- meal_history の days_ago は 0 が当日、1 が前日、2 が2日前を表す
- 出力はJSONのみ

入力データ:
%s`, string(b))

	raw, err := s.ai.GenerateRecommendation(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var res model.RecommendationResponse
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, fmt.Errorf("failed to parse recommendation response: %w; raw=%q", err, raw)
	}

	res.Error = nil
	return &res, nil
}

type mealHistoryForAI struct {
	Date     string   `json:"date"`
	DaysAgo  int      `json:"days_ago"`
	Content  []string `json:"content"`
	Calories float64  `json:"calories"`
	Protein  float64  `json:"protein"`
	Fat      float64  `json:"fat"`
	Carbs    float64  `json:"carbs"`
}

func buildRecommendationPayload(
	req model.RecommendationRequest,
	profile model.UserProfile,
	mealLogs []model.MealLog,
	activityLogs []model.ActivityLog,
) (map[string]any, error) {
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target_date: %w", err)
	}

	targetDay := dateOnly(targetDate)

	return map[string]any{
		"condition": req.Condition,
		"user_profile": profile,
		"activities": activityLogs,
		"meal_history": convertMealLogs(targetDay, mealLogs),
	}, nil
}

func convertMealLogs(targetDay time.Time, logs []model.MealLog) []mealHistoryForAI {
	out := make([]mealHistoryForAI, 0, len(logs))

	for _, log := range logs {
		d, err := time.Parse("2006-01-02", log.Date)
		if err != nil {
			continue
		}

		daysAgo := int(targetDay.Sub(dateOnly(d)).Hours() / 24)
		if daysAgo < 0 {
			daysAgo = 0
		}

		out = append(out, mealHistoryForAI{
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

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}