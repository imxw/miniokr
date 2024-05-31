package v1

import "time"

// UserRequest 结构体用于接收API请求参数
// type UserRequest struct {
// 	ID   string `form:"id" binding:"omitempty"`
// 	Name string `form:"name" binding:"omitempty"`
// }

// DepartmentDetails 结构体用于在API响应中包含部门信息
type DepartmentDetails struct {
	Name     string `json:"name"`
	IsLeader bool   `json:"isLeader"`
}

// UserResponse 结构体用于API响应
type UserResponse struct {
	UserID      string              `json:"uid"`
	Name        string              `json:"name"`
	Title       string              `json:"title"`
	Status      string              `json:"status"`
	Avatar      string              `json:"avatar"`
	JobNumber   string              `json:"jobNumber"`
	HiredDate   *time.Time          `json:"hiredDate,omitempty"` // 入职日期
	Roles       []string            `json:"roles"`
	Departments []DepartmentDetails `json:"departments"`
}

type TreeNode struct {
	Title    string      `json:"title"`
	Key      string      `json:"key"`
	Sort     int         `json:"sort"`
	Role     string      `json:"role,omitempty"`
	Position string      `json:"position,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
}
