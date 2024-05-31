package store

import (
	"sync"

	"gorm.io/gorm"

	"github.com/imxw/miniokr/internal/pkg/model"
)

var (
	once sync.Once
	// 全局变量，方便其它包直接调用已初始化好的 S 实例.
	S *datastore
)

// IStore 定义了 Store 层需要实现的方法.
type IStore interface {
	DB() *gorm.DB
	Users() UserStore
	Sync() SyncStorer
	AutoMigrate() error
	InitRoles() error
	InitUserRoles() error
}

var _ IStore = (*datastore)(nil)

// datastore 是 IStore 的一个具体实现.
type datastore struct {
	db *gorm.DB
}

// 确保 datastore 实现了 IStore 接口.
var _ IStore = (*datastore)(nil)

// NewStore 创建一个 IStore 类型的实例.
func NewStore(db *gorm.DB) *datastore {
	// 确保 S 只被初始化一次
	once.Do(func() {
		S = &datastore{db}
	})

	return S
}

// DB 返回存储在 datastore 中的 *gorm.DB.
func (ds *datastore) DB() *gorm.DB {
	return ds.db
}

// Users 返回一个实现了 UserStore 接口的实例.
func (ds *datastore) Users() UserStore {
	return newUsers(ds.db)
}

func (ds *datastore) Sync() SyncStorer {
	return NewSyncStore(ds.db)
}

// AutoMigrate 进行自动迁移
func (ds *datastore) AutoMigrate() error {
	// 按照依赖顺序迁移表

	if err := ds.db.AutoMigrate(&model.Department{}); err != nil {
		return err
	}

	if err := ds.db.AutoMigrate(&model.User{}); err != nil {
		return err
	}

	if err := ds.db.AutoMigrate(&model.UserDepartment{}); err != nil {
		return err
	}

	if err := ds.db.AutoMigrate(&model.Role{}); err != nil {
		return err
	}

	if err := ds.db.AutoMigrate(&model.UserRole{}); err != nil {
		return err
	}

	return nil
	// return ds.db.AutoMigrate(
	// 	&model.Department{},
	// 	&model.User{},
	// 	&model.UserDepartment{},
	// 	&model.Role{},
	// 	&model.UserRole{},
	// )
}

// InitRoles 初始化角色
func (ds *datastore) InitRoles() error {
	defaultRoles := []model.Role{
		{RoleName: "admin"},
		{RoleName: "leader"},
		{RoleName: "member"},
	}

	for _, role := range defaultRoles {
		if err := ds.db.Where("role_name = ?", role.RoleName).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

// InitUserRoles 初始化用户角色
func (ds *datastore) InitUserRoles() error {
	var userDepartments []model.UserDepartment
	if err := ds.db.Where("is_leader = ?", model.True).Find(&userDepartments).Error; err != nil {
		return err
	}

	var leaderRole model.Role
	if err := ds.db.Where("role_name = ?", "leader").First(&leaderRole).Error; err != nil {
		return err
	}

	for _, ud := range userDepartments {
		userRole := model.UserRole{
			UserID: ud.UserID,
			RoleID: leaderRole.RoleID,
		}
		if err := ds.db.Where("user_id = ? AND role_id = ?", ud.UserID, leaderRole.RoleID).FirstOrCreate(&userRole).Error; err != nil {
			return err
		}
	}
	return nil
}
