package models

import "gorm.io/gorm"

// UserProfile 用户扩展资料表。
// 使用一对一拆表，避免用户主表字段不断膨胀，便于后续平滑扩展。
type UserProfile struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`

	Nickname string `gorm:"type:varchar(64)" json:"nickname,omitempty"`
	Avatar   string `gorm:"type:varchar(512)" json:"avatar,omitempty"`
	Bio      string `gorm:"type:text" json:"bio,omitempty"`

	HeightCM           *float64 `gorm:"type:numeric(5,2)" json:"height_cm,omitempty"`
	WeightKG           *float64 `gorm:"type:numeric(5,2)" json:"weight_kg,omitempty"`
	Allergies          []string `gorm:"type:jsonb;serializer:json" json:"allergies,omitempty"`
	DietaryPreferences []string `gorm:"type:jsonb;serializer:json" json:"dietary_preferences,omitempty"`
	ExerciseExperience string   `gorm:"type:text" json:"exercise_experience,omitempty"`
	FitnessGoal        string   `gorm:"type:varchar(32)" json:"fitness_goal,omitempty"`
	MonthlyDietBudget  *int64   `gorm:"type:bigint" json:"monthly_diet_budget,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
