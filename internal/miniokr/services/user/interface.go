package user

import (
	"context"

	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

type Service interface {
	GetUserByID(context.Context, string) (*v1.UserResponse, error)
	GetUserByName(context.Context, string) (*v1.UserResponse, error)
	GetUserRolesByID(context.Context, string) ([]string, error)
	GetManagedUserIDs(context.Context, string) ([]string, error)
	GetUserDepartmentTree(context.Context, string) (*v1.TreeNode, error)
	GetCompanyDepartmentTree(context.Context) (*v1.TreeNode, error)
}
