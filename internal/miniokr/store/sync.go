package store

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/imxw/miniokr/internal/pkg/model"
)

type SyncStorer interface {
	PersistUsers(context.Context, []model.User) error
	PersistDepartments(context.Context, []model.Department) error
	PersistUserDepartments(context.Context, []model.UserDepartment) error
}

var _ SyncStorer = (*SyncStore)(nil)

type SyncStore struct {
	db *gorm.DB
}

func NewSyncStore(db *gorm.DB) *SyncStore {
	return &SyncStore{db: db}
}

func (s *SyncStore) PersistDepartments(ctx context.Context, departments []model.Department) error {
	// 实现持久化部门信息的逻辑
	for _, dept := range departments {
		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&dept).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncStore) PersistUsers(ctx context.Context, users []model.User) error {
	// 实现持久化用户信息的逻辑
	for _, user := range users {
		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&user).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncStore) PersistUserDepartments(ctx context.Context, userDepts []model.UserDepartment) error {
	// 实现持久化用户部门映射信息的逻辑
	for _, userDept := range userDepts {
		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&userDept).Error
		if err != nil {
			return err
		}
	}
	return nil
}
