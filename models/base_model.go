package models

import (
	"time"
)

// IEntity 泛型实体接口
type IEntity interface {
	GetID() uint
	TableName() string
}

// BaseModel 所有模型的基类
type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// 添加TableName的默认实现
func (b BaseModel) TableName() string {
	return ""
}

// 确保BaseModel实现IEntity接口
func (b BaseModel) GetID() uint {
	return b.ID
}
