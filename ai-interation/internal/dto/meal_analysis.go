package dto

type MealAnalysisResponse struct {
	Date     string   `json:"date"`
	Content  []string  `json:"content"`
	Calories float64   `json:"calories"`
	Protein  float64   `json:"protein"`
	Fat      float64   `json:"fat"`
	Carbs    float64   `json:"carbs"`
	Error    any      `json:"error"`
}