package models

// User 用户模型（示例）
type User struct {
	BaseModel
	Name     string `gorm:"size:100" json:"name"`
	Email    string `gorm:"uniqueIndex;size:255" json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"isActive"`
}

// TableName 获取表名
func (User) TableName() string {
	return "users"
}
