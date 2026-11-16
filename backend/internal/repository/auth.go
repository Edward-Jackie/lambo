package repository

import (
	"context"
	"time"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthRepository struct{ *Repository }

func NewAuthRepository(repository *Repository) *AuthRepository { return &AuthRepository{repository} }

func (r *AuthRepository) FindAccount(ctx context.Context, username string) (*model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&account).Error
	return &account, err
}

func (r *AuthRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "username"}}, DoNothing: true}).Create(account).Error
}

func (r *AuthRepository) SaveSession(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at <= ?", time.Now().UTC()).Delete(&model.Session{}).Error; err != nil {
			return err
		}
		return tx.Create(session).Error
	})
}

func (r *AuthRepository) SessionAccount(ctx context.Context, tokenHash string) (*model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).Table("lambo_accounts").Select("lambo_accounts.*").
		Joins("JOIN lambo_sessions ON lambo_sessions.account_id = lambo_accounts.id").
		Where("lambo_sessions.token_hash = ? AND lambo_sessions.expires_at > ? AND lambo_accounts.enabled = ?", tokenHash, time.Now().UTC(), true).
		Take(&account).Error
	return &account, err
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&model.Session{}).Error
}
