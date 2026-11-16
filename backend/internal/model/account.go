package model

import "time"

const (
	RoleAdmin       = "admin"
	RoleReader      = "reader"
	SessionLifetime = 8 * time.Hour
)

// Account 是本地管理后台账号，与既有用户示例表隔离。
type Account struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"size:64;not null;uniqueIndex"`
	PasswordHash string `gorm:"size:100;not null" json:"-"`
	RealName     string `gorm:"size:64;not null"`
	Role         string `gorm:"size:32;not null"`
	Enabled      bool   `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (*Account) TableName() string { return "lambo_accounts" }

// Session 仅持久化令牌摘要，登出后立即撤销服务端会话。
type Session struct {
	TokenHash string    `gorm:"size:64;primaryKey" json:"-"`
	AccountID uint64    `gorm:"not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}

func (*Session) TableName() string { return "lambo_sessions" }
