package models

import (
	"time"

	"gorm.io/gorm"
)

// MealRecord 用户饮食记录。
type MealRecord struct {
	gorm.Model
	UserID uint      `gorm:"index;not null" json:"user_id"`
	Date   time.Time `gorm:"type:date;index;not null" json:"date"`

	Type    string `gorm:"type:varchar(16);not null;index" json:"type"`
	Content string `gorm:"type:text;not null" json:"content"`

	Calories *int     `gorm:"type:int" json:"calories,omitempty"`
	Protein  *float64 `gorm:"type:numeric(6,2)" json:"protein,omitempty"`
	Carbs    *float64 `gorm:"type:numeric(6,2)" json:"carbs,omitempty"`
	Fat      *float64 `gorm:"type:numeric(6,2)" json:"fat,omitempty"`

	Source           string `gorm:"type:varchar(64)" json:"source,omitempty"`
	RecommendationID string `gorm:"type:varchar(128)" json:"recommendation_id,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
