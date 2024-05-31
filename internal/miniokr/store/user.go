package store

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/imxw/miniokr/internal/pkg/model"
)

type UserStore interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByName(ctx context.Context, username string) (*model.User, error)
	GetUserDepartments(ctx context.Context, userID string) ([]model.UserDepartment, error)
	GetDepartmentByID(ctx context.Context, departmentID int) (*model.Department, error)
	GetUserRolesByID(ctx context.Context, userID string) ([]string, error)
	GetManagedDepartments(ctx context.Context, userID string) ([]int, error)
	GetUserIDsByDepartmentIDs(ctx context.Context, departmentIDs []int) ([]string, error)
	GetDepartmentsByParentID(ctx context.Context, parentID int) ([]model.Department, error)
	GetUsersByDepartmentID(ctx context.Context, departmentID int) ([]model.User, error)
	GetUserDepartment(ctx context.Context, userID string, departmentID int) (*model.UserDepartment, error)
	GetParentDepartment(ctx context.Context, departmentID int) (*model.Department, error)
}

// UserStore 接口的实现.
type users struct {
	db *gorm.DB
}

// 确保 users 实现了 UserStore 接口.
var _ UserStore = (*users)(nil)

func newUsers(db *gorm.DB) *users {
	return &users{db}
}

func (s *users) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	if result := s.db.Preload("UserDepartments").Preload("UserDepartments.Department").First(&user, "user_id = ?", id); result.Error != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *users) GetUserByName(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if result := s.db.Preload("UserDepartments").Preload("UserDepartments.Department").First(&user, "name = ?", username); result.Error != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// GetUserRolesByID 根据用户ID获取用户的角色
func (s *users) GetUserRolesByID(ctx context.Context, userID string) ([]string, error) {
	var roles []model.Role
	if err := s.db.Joins("JOIN user_roles ON user_roles.role_id = roles.role_id").
		Where("user_roles.user_id = ?", userID).Find(&roles).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	if len(roles) == 0 {
		return []string{"member"}, nil
	}
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.RoleName
	}
	return roleNames, nil
}

func (s *users) GetUserDepartments(ctx context.Context, userID string) ([]model.UserDepartment, error) {
	var userDepartments []model.UserDepartment
	if result := s.db.Where("user_id = ?", userID).Find(&userDepartments); result.Error != nil {
		return nil, result.Error
	}
	return userDepartments, nil
}

func (s *users) GetDepartmentByID(ctx context.Context, departmentID int) (*model.Department, error) {
	var department model.Department
	if result := s.db.First(&department, "department_id = ?", departmentID); result.Error != nil {
		return nil, errors.New("department not found")
	}
	return &department, nil
}

func (s *users) GetManagedDepartments(ctx context.Context, userID string) ([]int, error) {
	var managedDepartments []int
	if err := s.db.Table("user_departments").Where("user_id = ? AND is_leader = ?", userID, true).Pluck("department_id", &managedDepartments).Error; err != nil {
		return nil, err
	}
	if len(managedDepartments) == 0 {
		return nil, errors.New("user does not manage any departments")
	}
	return managedDepartments, nil
}

func (s *users) GetUserIDsByDepartmentIDs(ctx context.Context, departmentIDs []int) ([]string, error) {
	var userIDs []string
	if err := s.db.Table("user_departments").Where("department_id IN (?)", departmentIDs).
		Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	// 使用map去重
	userIDMap := make(map[string]struct{})
	for _, id := range userIDs {
		userIDMap[id] = struct{}{}
	}

	uniqueUserIDs := make([]string, 0, len(userIDMap))
	for id := range userIDMap {
		uniqueUserIDs = append(uniqueUserIDs, id)
	}

	return uniqueUserIDs, nil
}

func (s *users) GetDepartmentsByParentID(ctx context.Context, parentID int) ([]model.Department, error) {
	var depts []model.Department
	if err := s.db.Where("parent_id = ?", parentID).Order("sort").Find(&depts).Error; err != nil {
		return nil, err
	}
	return depts, nil
}

func (s *users) GetUsersByDepartmentID(ctx context.Context, departmentID int) ([]model.User, error) {
	var users []model.User
	if err := s.db.Joins("JOIN user_departments ON user_departments.user_id = users.user_id").
		Where("user_departments.department_id = ?", departmentID).
		Order("users.sort").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *users) GetUserDepartment(ctx context.Context, userID string, departmentID int) (*model.UserDepartment, error) {
	var userDept model.UserDepartment
	if err := s.db.Where("user_id = ? AND department_id = ?", userID, departmentID).First(&userDept).Error; err != nil {
		return nil, err
	}
	return &userDept, nil
}

func (s *users) GetParentDepartment(ctx context.Context, departmentID int) (*model.Department, error) {
	dept, err := s.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	if dept.ParentID == nil || *dept.ParentID == 1 {
		return &model.Department{}, nil
	}

	parentDept, err := s.GetDepartmentByID(ctx, *dept.ParentID)
	if err != nil {
		return nil, err
	}

	return parentDept, nil
}
