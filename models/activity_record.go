package models

import (
	"time"

	"gorm.io/gorm"
)

// ActivityRecord 用户行程/活动记录。
type ActivityRecord struct {
	gorm.Model
	UserID uint      `gorm:"index;not null" json:"user_id"`
	Date   time.Time `gorm:"type:date;index;not null" json:"date"`

	Title string `gorm:"type:varchar(128);not null" json:"title"`

	StartTime string `gorm:"type:varchar(8);not null" json:"start_time"`
	EndTime   string `gorm:"type:varchar(8);not null" json:"end_time"`

	Intensity string `gorm:"type:varchar(16);index" json:"intensity,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
