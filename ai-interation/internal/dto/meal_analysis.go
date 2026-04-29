package dto

type MealItem struct {
	Name string `json:"name"`
}

type TotalNutrition struct {
	Calories         float64 `json:"calories"`
	Protein          float64 `json:"protein"`
	Fat              float64 `json:"fat"`
	Carbohydrates    float64 `json:"carbohydrates"`
	VegetablesAmount string  `json:"vegetables_amount"`
}

type MealAnalysisResponse struct {
	Items         []MealItem      `json:"items"`
	TotalNutrition TotalNutrition  `json:"total_nutrition"`
	Error         any             `json:"error"`
}