package models

import (
	"time"

	"gorm.io/gorm"
)

// UserSession 用户登录会话表。
// 目标：将“已登录用户状态”持久化到数据库，支持服务重启后状态可追踪，
// 并为后续多端登录管理、风控、审计、在线设备管理打基础。
type UserSession struct {
	gorm.Model
	UserID uint `gorm:"index;not null" json:"user_id"`

	// SessionID 为本次登录会话 ID。
	SessionID string `gorm:"type:varchar(64);uniqueIndex;not null" json:"session_id"`

	// AccessTokenJTI / RefreshTokenJTI 对应 JWT 的 jti。
	AccessTokenJTI  string `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	RefreshTokenJTI string `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`

	AccessExpiresAt  time.Time `gorm:"index;not null" json:"access_expires_at"`
	RefreshExpiresAt time.Time `gorm:"index;not null" json:"refresh_expires_at"`
	LastSeenAt       time.Time `gorm:"index;not null" json:"last_seen_at"`

	// IsRevoked 为 true 表示该会话已失效（主动登出或风控封禁）。
	IsRevoked bool       `gorm:"index;not null;default:false" json:"is_revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	ClientIP  string `gorm:"type:varchar(64)" json:"client_ip,omitempty"`
	UserAgent string `gorm:"type:varchar(512)" json:"user_agent,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
