package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
)

const iamSequenceUpgradeKey = "rbac-sequences-v2"

// UpgradeIAM 前移旧库序列，保留既有 ID、授权和已预留的序列值。
func (r *AuthRepository) UpgradeIAM(ctx context.Context) error {
	var hasMarkers bool
	if err := r.db.WithContext(ctx).Raw("SELECT to_regclass(?) IS NOT NULL", (model.IAMMigration{}).TableName()).Scan(&hasMarkers).Error; err != nil || !hasMarkers {
		return err
	}
	return r.withIAMLock(ctx, func(tx *gorm.DB) error {
		initialized, err := iamMarked(tx, iamBootstrapKey)
		if err != nil || !initialized {
			return err
		}
		return upgradeIAMSequences(tx)
	})
}

func upgradeIAMSequences(tx *gorm.DB) error {
	upgraded, err := iamMarked(tx, iamSequenceUpgradeKey)
	if err != nil || upgraded {
		return err
	}
	for _, table := range []interface{ TableName() string }{&model.SystemRole{}, &model.SystemMenu{}} {
		if err := advanceIAMSequence(tx, table); err != nil {
			return fmt.Errorf("升级 %s 的自增序列：%w", table.TableName(), err)
		}
	}
	return tx.Create(&model.IAMMigration{Key: iamSequenceUpgradeKey}).Error
}

func advanceIAMSequence(tx *gorm.DB, table interface{ TableName() string }) error {
	var sequence *string
	if err := tx.Raw("SELECT pg_get_serial_sequence(?, 'id')", table.TableName()).Scan(&sequence).Error; err != nil {
		return err
	}
	if sequence == nil {
		return fmt.Errorf("缺少与主键关联的 PostgreSQL 序列")
	}
	var largestID int64
	if err := tx.Model(table).Select("COALESCE(MAX(id), 0)").Scan(&largestID).Error; err != nil {
		return err
	}
	var state struct {
		LastValue int64
		IsCalled  bool
	}
	// pg_get_serial_sequence 返回按 PostgreSQL 规则引用的限定名称。
	if err := tx.Raw("SELECT last_value, is_called FROM " + *sequence).Scan(&state).Error; err != nil {
		return err
	}
	if state.LastValue > largestID || (state.IsCalled && state.LastValue == largestID) {
		return nil
	}
	if largestID == math.MaxInt64 {
		return fmt.Errorf("已用尽 PostgreSQL 主键序列")
	}
	return tx.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH %d", *sequence, largestID+1)).Error
}

// CheckUsernameUpgrade 在建立唯一索引前明确报告旧库的重名数据。
func CheckUsernameUpgrade(ctx context.Context, db *gorm.DB) error {
	db = db.WithContext(ctx)
	var check bool
	if err := db.Raw("SELECT to_regclass(?) IS NOT NULL AND to_regclass(?) IS NULL", (&model.SystemUser{}).TableName(), model.AccountUsernameIndex).Scan(&check).Error; err != nil || !check {
		return err
	}
	var duplicate bool
	groups := db.Model(&model.SystemUser{}).Select("username").Group("username").Having("count(*) > 1")
	if err := db.Raw("SELECT EXISTS (?)", groups).Scan(&duplicate).Error; err != nil {
		return err
	}
	if duplicate {
		return iamConflict("旧库存在重复用户名，请核对账号归属并处理重名后重试迁移")
	}
	return nil
}
