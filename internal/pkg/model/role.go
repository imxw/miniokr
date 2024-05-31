package model

// Role 结构体
type Role struct {
	RoleID   uint   `gorm:"primaryKey;autoIncrement"`
	RoleName string `gorm:"size:50;not null;unique"`
}

// UserRole 结构体
type UserRole struct {
	UserID string `gorm:"size:255;not null"`
	RoleID uint   `gorm:"not null"`
}
