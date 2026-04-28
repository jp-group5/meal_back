package models

import "gorm.io/gorm"

// User 用户基础模型。
// 设计说明：
// 1. Username 作为主要登录标识，唯一索引保证不可重复。
// 2. Email 保留为可选字段，便于后续找回密码、通知等场景。
// 3. PasswordHash 仅存储密码哈希值，绝不存储明文密码。
// 4. Profile 与 Sessions 为扩展关系，支持后续业务演进。
type User struct {
	gorm.Model
	Username     string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        *string       `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	PasswordHash string        `gorm:"type:text;not null" json:"-"`
	Profile      *UserProfile  `json:"profile,omitempty"`
	Sessions     []UserSession `json:"-"`
}
