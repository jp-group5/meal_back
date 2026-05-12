package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"ai-interation/internal/dto"
	"ai-interation/internal/port"
)

type RecommendationUsecase struct {
	aiClient   port.AIClient
	repository port.Repository
}

func NewRecommendationUsecase(aiClient port.AIClient, repository port.Repository) *RecommendationUsecase {
	return &RecommendationUsecase{
		aiClient:   aiClient,
		repository: repository,
	}
}

type dietaryHistoryForAI struct {
	FoodName  string   `json:"food_name"`
	MealType  string   `json:"meal_type"`
	DaysAgo   int      `json:"days_ago"`
	Tags      []string `json:"tags"`
}

type todayContextForAI struct {
	CurrentTime               string `json:"current_time"`
	TotalBurnedCaloriesEstimate int   `json:"total_burned_calories_estimate"`
	Activities                []port.ActivityLog `json:"activities"`
}

type recommendationPayloadForAI struct {
	Condition       string                 `json:"condition"`
	UserMeta        port.UserProfile       `json:"user_meta"`
	TodayContext    todayContextForAI      `json:"today_context"`
	DietaryHistory  []dietaryHistoryForAI  `json:"dietary_history"`
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

	dietaryHistory, err := u.repository.GetDietaryHistory(ctx, req.UserID, req.TargetDate)
	if err != nil {
		return nil, err
	}

	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target_date: %w", err)
	}
	targetDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)

	convertedHistory := convertDietaryHistory(targetDay, dietaryHistory)

	// 必要ならここで今日の消費カロリーを別計算 or DB取得に差し替え
	estimatedBurned := 0
	for _, act := range activityLogs {
		_ = act
	}

	payload := recommendationPayloadForAI{
		Condition: req.Condition,
		UserMeta:  profile,
		TodayContext: todayContextForAI{
			CurrentTime:               targetDate.Format(time.RFC3339),
			TotalBurnedCaloriesEstimate: estimatedBurned,
			Activities:                activityLogs,
		},
		DietaryHistory: convertedHistory,
	}

	// mealLogs も必要なら prompt に追加できます
	_ = mealLogs

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`あなたは栄養バランスを考慮した献立提案の専門家です。
以下のデータをもとに、おすすめ献立を3件、JSONのみで返してください。

制約:
- アレルギーは必ず避ける
- preferences を考慮する
- health_goal に沿う
- recent な dietary_history を優先して偏りを避ける
- dietary_history の days_ago は 0 が当日、1 が前日、2 が2日前
- menu_name及びreasonは日本語で返す
- reasonは50文字以下

データ:
%s

返却形式:
{
  "recommendations": [
    { "menu_name": "...", "reason": "..." }
  ],
  "error": null
}`, string(b))

	raw, err := u.aiClient.GenerateRecommendation(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var res dto.RecommendationResponse
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, err
	}

	res.Error = nil
	return &res, nil
}

func convertDietaryHistory(targetDay time.Time, items []port.DietaryHistory) []dietaryHistoryForAI {
	out := make([]dietaryHistoryForAI, 0, len(items))

	sort.Slice(items, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, items[i].Timestamp)
		tj, errJ := time.Parse(time.RFC3339, items[j].Timestamp)
		if errI != nil || errJ != nil {
			return i < j
		}
		return ti.After(tj) // 新しい順
	})

	for _, item := range items {
		t, err := time.Parse(time.RFC3339, item.Timestamp)
		if err != nil {
			continue
		}

		historyDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		daysAgo := int(targetDay.Sub(historyDay).Hours() / 24)
		if daysAgo < 0 {
			daysAgo = 0
		}

		out = append(out, dietaryHistoryForAI{
			FoodName: item.FoodName,
			MealType: item.MealType,
			DaysAgo:  daysAgo,
			Tags:     item.Tags,
		})
	}

	return out
}