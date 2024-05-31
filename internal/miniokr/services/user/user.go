package user

import (
	"context"
	"errors"
	"strings"

	"github.com/imxw/miniokr/internal/miniokr/store"
	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

var _ Service = (*UserService)(nil)

type UserService struct {
	store store.UserStore
}

func NewUserService(store store.UserStore) *UserService {
	return &UserService{store: store}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*v1.UserResponse, error) {
	return s.getUser(ctx, "id", id)
}

func (s *UserService) GetUserByName(ctx context.Context, username string) (*v1.UserResponse, error) {
	return s.getUser(ctx, "name", username)
}

func (s *UserService) getUser(ctx context.Context, field, value string) (*v1.UserResponse, error) {
	var user *model.User
	var err error
	if field == "id" {
		user, err = s.store.GetUserByID(ctx, value)
	} else {
		user, err = s.store.GetUserByName(ctx, value)
	}
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 获取用户的角色信息
	roles, err := s.store.GetUserRolesByID(ctx, user.UserID)
	if err != nil {
		return nil, errors.New("failed to get user role")
	}

	departmentMap := make(map[int]bool)
	departmentDetails := make([]v1.DepartmentDetails, 0)
	for _, ud := range user.UserDepartments {

		if ud.DepartmentID == 1 {
			// 如果用户在公司级别，部门信息为空
			departmentDetails = []v1.DepartmentDetails{}
			break
		} else {
			departments, err := s.getDepartmentHierarchy(ctx, ud.DepartmentID)
			if err != nil {
				return nil, err
			}
			fullDepartmentName := s.getFullDepartmentName(departments)
			for i, dept := range departments {
				if !departmentMap[dept.DepartmentID] {
					departmentDetails = append(departmentDetails, v1.DepartmentDetails{
						Name:     dept.Name,
						IsLeader: ud.IsLeader == model.True && dept.DepartmentID == ud.DepartmentID,
					})
					departmentMap[dept.DepartmentID] = true
				}
				if i > 0 { // 如果不是最高级部门，则拼接父部门名称
					departmentDetails[len(departmentDetails)-1].Name = fullDepartmentName
				}
			}
		}
	}

	return &v1.UserResponse{
		UserID:      user.UserID,
		Name:        user.Name,
		Title:       user.Title,
		Status:      user.Status,
		Avatar:      user.Avatar,
		HiredDate:   user.HiredDate,
		JobNumber:   user.JobNumber,
		Roles:       roles,
		Departments: departmentDetails,
	}, nil
}

func (s *UserService) getDepartmentHierarchy(ctx context.Context, departmentID int) ([]model.Department, error) {
	var departments []model.Department
	currentDeptID := departmentID

	for currentDeptID != 0 {
		dept, err := s.store.GetDepartmentByID(ctx, currentDeptID)
		if err != nil {
			return nil, errors.New("department not found")
		}
		departments = append([]model.Department{*dept}, departments...)
		if dept.ParentID == nil || *dept.ParentID == 1 {
			break
		}
		currentDeptID = *dept.ParentID
	}
	return departments, nil
}

func (s *UserService) getFullDepartmentName(departments []model.Department) string {
	var names []string
	for _, dept := range departments {
		names = append(names, dept.Name)
	}
	return strings.Join(names, "-")
}

func (s *UserService) GetManagedUserIDs(ctx context.Context, userID string) ([]string, error) {

	managedDepartments, err := s.store.GetManagedDepartments(ctx, userID)
	if err != nil {
		return nil, err
	}

	userIDs, err := s.store.GetUserIDsByDepartmentIDs(ctx, managedDepartments)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (s *UserService) GetUserRolesByID(ctx context.Context, userID string) ([]string, error) {
	return s.store.GetUserRolesByID(ctx, userID)
}
