package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Department 结构体
type Department struct {
	DepartmentID int       `gorm:"primaryKey"`
	Name         string    `gorm:"size:255;not null"`
	ParentID     *int      `gorm:"index"`
	Sort         int       `gorm:"index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// User 结构体
type User struct {
	UserID          string           `gorm:"primaryKey;size:255"`
	Name            string           `gorm:"size:255;not null"`
	Title           string           `gorm:"size:255"`
	Status          string           `gorm:"size:100"`
	Mobile          string           `gorm:"size:20"`
	Avatar          string           `gorm:"size:255"`
	JobNumber       string           `gorm:"size:50"`
	Sort            int              `gorm:"index"`
	HiredDate       *time.Time       `gorm:"type:timestamp"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"`
	UserDepartments []UserDepartment `gorm:"foreignKey:UserID;references:UserID"`
}

// UserDepartment 结构体
type UserDepartment struct {
	UserID       string     `gorm:"primaryKey;size:255"`
	DepartmentID int        `gorm:"primaryKey"`
	IsLeader     Bool       `gorm:"type:tinyint(1)"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
	Department   Department `gorm:"foreignKey:DepartmentID;references:DepartmentID"`
}

// Bool 类型表示布尔值
type Bool int8

const (
	False Bool = 0
	True  Bool = 1
)

// 实现 driver.Valuer 接口
func (b Bool) Value() (driver.Value, error) {
	return int64(b), nil
}

// 实现 sql.Scanner 接口
func (b *Bool) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		*b = Bool(v)
	case bool:
		if v {
			*b = True
		} else {
			*b = False
		}
	default:
		return fmt.Errorf("unsupported Scan value for Bool: %v", value)
	}
	return nil
}
