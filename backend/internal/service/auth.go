package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/repository"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUnauthenticated = errors.New("账号密码错误或会话已过期")

type AuthService struct {
	repository *repository.AuthRepository
	dummyHash  []byte
}

func NewAuthService(repo *repository.AuthRepository, conf *viper.Viper) (*AuthService, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("仅用于等时比较的无效账号"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	svc := &AuthService{repository: repo, dummyHash: dummy}
	if conf.GetBool("bootstrap.enabled") {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for _, seed := range []struct{ key, role, name string }{{"admin", model.RoleAdmin, "管理员"}, {"reader", model.RoleReader, "只读用户"}} {
			username := strings.TrimSpace(conf.GetString("bootstrap." + seed.key + "_username"))
			password := conf.GetString("bootstrap." + seed.key + "_password")
			if username == "" {
				continue
			}
			_, err := repo.FindAccount(ctx, username)
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			if len(username) > 64 || len(password) < 12 || len(password) > 72 {
				return nil, fmt.Errorf("初始化账号 %s 需要有效用户名及 12—72 字节密码", seed.key)
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return nil, err
			}
			account := &model.Account{Username: username, PasswordHash: string(hash), RealName: seed.name, Role: seed.role, Enabled: true}
			if err := repo.CreateAccount(ctx, account); err != nil {
				return nil, fmt.Errorf("初始化账号失败：%w", err)
			}
		}
	}
	if conf.GetBool("data.auto_migrate") || conf.GetBool("bootstrap.enabled") {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := repo.InitializeIAM(ctx); err != nil {
			return nil, err
		}
	}
	return svc, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	if len(username) > 64 || len(password) > 72 || username == "" || password == "" {
		return "", ErrUnauthenticated
	}
	account, err := s.repository.FindAccount(ctx, username)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return "", ErrUnauthenticated
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil || !account.Enabled {
		return "", ErrUnauthenticated
	}
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(value)
	session := &model.Session{TokenHash: hashToken(token), AccountID: account.ID, ExpiresAt: time.Now().UTC().Add(model.SessionLifetime)}
	if err := s.repository.SaveSession(ctx, session); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*model.Account, error) {
	if len(token) != 43 {
		return nil, ErrUnauthenticated
	}
	account, err := s.repository.SessionAccount(ctx, hashToken(token))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUnauthenticated
	}
	return account, err
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if len(token) != 43 {
		return nil
	}
	return s.repository.DeleteSession(ctx, hashToken(token))
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) ManagementOptions(ctx context.Context) (map[string]interface{}, error) {
	return s.repository.ManagementOptions(ctx)
}
