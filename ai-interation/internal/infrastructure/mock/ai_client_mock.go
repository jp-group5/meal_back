package mock

import "context"

type AIClientMock struct{}

func NewAIClientMock() *AIClientMock {
	return &AIClientMock{}
}

func (a *AIClientMock) AnalyzeMeal(ctx context.Context, imageBytes []byte) (string, error) {
	return `{
		"date": "2026-04-28",
		"content": ["焼き鮭", "みそ汁", "漬物", "白米"],
		"calories": 490,
		"protein": 27.4,
		"fat": 10.8,
		"carbs": 62.0,
		"error": null
	}`, nil
}

func (a *AIClientMock) GenerateRecommendation(ctx context.Context, prompt string) (string, error) {
	return `{
		"recommendations": [
			{
				"menu_name": "鶏むね肉のサラダ定食",
				"calories": 520,
				"protein": 42.0,
				"fat": 18.0,
				"carbs": 28.0,
				"reason": "たんぱく質を確保しつつ、脂質を抑えやすいため"
			},
			{
				"menu_name": "鮭と野菜の蒸し料理",
				"calories": 460,
				"protein": 31.0,
				"fat": 14.0,
				"carbs": 35.0,
				"reason": "野菜量を増やしやすく、栄養バランスを整えやすいため"
			},
			{
				"menu_name": "豆腐と野菜のうどん",
				"calories": 410,
				"protein": 18.0,
				"fat": 9.0,
				"carbs": 58.0,
				"reason": "消化しやすく、簡単に用意できるため"
			}
		],
		"error": null
	}`, nil
}