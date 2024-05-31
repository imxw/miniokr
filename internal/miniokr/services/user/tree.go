package user

import (
	"context"
	"fmt"
	"sort"

	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

const RootDeptID = 1
const CompanyName = "托普汇智(北京)科技有限公司"

func (s *UserService) GetCompanyDepartmentTree(ctx context.Context) (*v1.TreeNode, error) {
	return s.getDepartmentTree(ctx, RootDeptID)
}

func (s *UserService) GetUserDepartmentTree(ctx context.Context, userID string) (*v1.TreeNode, error) {
	userDepts, err := s.store.GetUserDepartments(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(userDepts) == 0 {
		return nil, fmt.Errorf("user %s is not part of any department", userID)
	}

	var leaderTree []*v1.TreeNode

	if len(userDepts) == 1 {
		subTree, err := s.getDepartmentTree(ctx, userDepts[0].DepartmentID)
		if err != nil {
			return nil, err
		}
		leaderTree = append(leaderTree, subTree)
	} else {
		filteredDepts := s.filterParentDepartments(ctx, userDepts)

		for _, deptID := range filteredDepts {
			subTree, err := s.getDepartmentTree(ctx, deptID)
			if err != nil {
				return nil, err
			}
			leaderTree = append(leaderTree, subTree)
		}
	}

	if s.getDeptIDFromKey(leaderTree[0].Key) == RootDeptID {
		return leaderTree[0], nil
	}

	return s.withCompanyNode(leaderTree), nil
}

func (s *UserService) getDepartmentTree(ctx context.Context, deptID int) (*v1.TreeNode, error) {
	if deptID == RootDeptID {
		return s.getCompanyTree(ctx)
	}

	dept, err := s.store.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return nil, err
	}

	node, err := s.buildTreeNode(ctx, *dept)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *UserService) getCompanyTree(ctx context.Context) (*v1.TreeNode, error) {
	var tree []*v1.TreeNode

	depts, err := s.store.GetDepartmentsByParentID(ctx, RootDeptID)
	if err != nil {
		return nil, err
	}

	userNodes, err := s.getDepartmentUsersAsTreeNodes(ctx, RootDeptID)
	if err != nil {
		return nil, err
	}

	for _, dept := range depts {
		node, err := s.buildTreeNode(ctx, dept)
		if err != nil {
			return nil, err
		}
		tree = append(tree, node)
	}

	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Sort < tree[j].Sort
	})
	sort.Slice(userNodes, func(i, j int) bool {
		return userNodes[i].Sort < userNodes[j].Sort
	})

	tree = append(tree, userNodes...)

	return s.withCompanyNode(tree), nil
}

func (s *UserService) buildTreeNode(ctx context.Context, dept model.Department) (*v1.TreeNode, error) {
	children, err := s.getDepartmentChildren(ctx, dept.DepartmentID)
	if err != nil {
		return nil, err
	}

	userNodes, err := s.getDepartmentUsersAsTreeNodes(ctx, dept.DepartmentID)
	if err != nil {
		return nil, err
	}

	sort.Slice(children, func(i, j int) bool {
		return children[i].Sort < children[j].Sort
	})
	sort.Slice(userNodes, func(i, j int) bool {
		return userNodes[i].Sort < userNodes[j].Sort
	})

	return &v1.TreeNode{
		Title:    dept.Name,
		Key:      fmt.Sprintf("dept-%d", dept.DepartmentID),
		Sort:     dept.Sort,
		Children: append(children, userNodes...),
	}, nil
}

func (s *UserService) getDepartmentChildren(ctx context.Context, deptID int) ([]*v1.TreeNode, error) {
	var children []*v1.TreeNode
	depts, err := s.store.GetDepartmentsByParentID(ctx, deptID)
	if err != nil {
		return nil, err
	}

	for _, dept := range depts {
		node, err := s.buildTreeNode(ctx, dept)
		if err != nil {
			return nil, err
		}
		children = append(children, node)
	}

	return children, nil
}

func (s *UserService) getDepartmentUsersAsTreeNodes(ctx context.Context, departmentID int) ([]*v1.TreeNode, error) {
	users, err := s.store.GetUsersByDepartmentID(ctx, departmentID)
	if err != nil {
		return nil, err
	}

	var userNodes []*v1.TreeNode
	for _, user := range users {
		var role string
		if userDept, err := s.store.GetUserDepartment(ctx, user.UserID, departmentID); err == nil && userDept.IsLeader == model.True {
			role = "leader"
		}

		userNodes = append(userNodes, &v1.TreeNode{
			Title:    user.Name,
			Key:      fmt.Sprintf("%d-%s", departmentID, user.UserID),
			Sort:     user.Sort,
			Role:     role,
			Position: user.Title,
		})
	}

	return userNodes, nil
}

func (s *UserService) filterParentDepartments(ctx context.Context, userDepts []model.UserDepartment) []int {
	departmentSet := make(map[int]struct{})
	for _, userDept := range userDepts {
		departmentSet[userDept.DepartmentID] = struct{}{}
	}

	filtered := make(map[int]struct{})
	for deptA := range departmentSet {
		shouldAdd := false
		for deptB := range departmentSet {
			if s.isParentDepartment(ctx, deptA, deptB) {
				shouldAdd = true
				continue
			}
		}
		if shouldAdd {
			filtered[deptA] = struct{}{}
		}
	}

	var result []int
	for deptID := range filtered {
		result = append(result, deptID)
	}

	return result
}

func (s *UserService) isParentDepartment(ctx context.Context, deptA, deptB int) bool {
	parent, err := s.store.GetParentDepartment(ctx, deptB)
	if err != nil {
		return false
	}
	if parent.DepartmentID == deptA {
		return true
	}
	return false
}

func (s *UserService) getDeptIDFromKey(key string) int {
	var deptID int
	fmt.Sscanf(key, "dept-%d", &deptID)
	return deptID
}

func (s *UserService) withCompanyNode(treeNode []*v1.TreeNode) *v1.TreeNode {
	return &v1.TreeNode{
		Title:    CompanyName,
		Key:      fmt.Sprintf("dept-%d", RootDeptID),
		Sort:     0,
		Children: treeNode,
	}
}
