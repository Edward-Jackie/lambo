package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"gorm.io/gorm"
	"regexp"
	"time"
)

const iamWriteLock int64 = 1279345218
const maxIAMMenus = 1000

var iamCodePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)
var iamPathPattern = regexp.MustCompile(`^/[a-zA-Z0-9/_-]+$`)

func textValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func iamInvalid(message string) error  { return fmt.Errorf("%w: %s", model.ErrIAMInvalid, message) }
func iamConflict(message string) error { return fmt.Errorf("%w: %s", model.ErrIAMConflict, message) }

func iamDBError(err error) error {
	var pg interface{ SQLState() string }
	if errors.As(err, &pg) && (pg.SQLState() == "23505" || pg.SQLState() == "23503") {
		return iamConflict("编码、用户名或路由重复，或记录仍有关联")
	}
	return err
}

// 权限配置低频写入使用数据库事务锁，使授权、最后管理员检查和会话撤销成为同一事务。
func (r *Repository) iamWrite(ctx context.Context, change func(*gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return iamDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", iamWriteLock).Error; err != nil {
			return err
		}
		if model.ActorID(ctx) == 0 {
			return model.ErrIAMForbidden
		}
		actor := model.Account{}
		if err := tx.Where("id=? AND enabled=?", model.ActorID(ctx), true).First(&actor).Error; err != nil {
			return model.ErrIAMForbidden
		}
		if err := loadAccess(tx, &actor); err != nil {
			return err
		}
		required := model.RequiredCode(ctx)
		owned := false
		for _, code := range actor.Codes {
			if code == required {
				owned = true
			}
		}
		if required == "" || !owned {
			return model.ErrIAMForbidden
		}
		if err := change(tx); err != nil {
			return err
		}
		return ensureManager(tx)
	}))
}

func uniqueIDs(values []uint64) ([]uint64, error) {
	seen := map[uint64]bool{}
	out := make([]uint64, 0, len(values))
	for _, id := range values {
		if id == 0 || id > 9007199254740991 || seen[id] {
			return nil, iamInvalid("关联 ID 无效或重复")
		}
		seen[id] = true
		out = append(out, id)
	}
	return out, nil
}

func revokeAccount(tx *gorm.DB, id uint64) error {
	return tx.Where("account_id=?", id).Delete(&model.Session{}).Error
}
func revokeRole(tx *gorm.DB, id uint64) error {
	return tx.Where("account_id IN (?)", tx.Model(&model.AccountRole{}).Select("account_id").Where("role_id=?", id)).Delete(&model.Session{}).Error
}

func revokeMenu(tx *gorm.DB, id uint64) error {
	affected := tx.Table("lambo_account_roles ar").Select("ar.account_id").Joins("JOIN lambo_role_menus rm ON rm.role_id=ar.role_id").Where("rm.menu_id=?", id)
	return tx.Where("account_id IN (?)", affected).Delete(&model.Session{}).Error
}
