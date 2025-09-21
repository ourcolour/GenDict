package services

import (
	"context"
	"goDict/dao"
	"goDict/models"
	"gorm.io/gorm"
)

// UserService 用户服务（可扩展自定义方法）
type UserService struct {
	BaseService[models.User]
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	userDao := dao.NewBaseDao[models.User](db)
	baseService := NewBaseService[models.User](userDao)
	return &UserService{BaseService: *baseService}
}

// CustomQueryByAge 自定义查询方法示例
func (s *UserService) CustomQueryByAge(ctx context.Context, minAge, maxAge int) ([]*models.User, error) {
	var users []*models.User
	db := s.dao.GetDb()

	result := db.WithContext(ctx).Table(s.dao.GetTableName()).Where("age BETWEEN ? AND ?", minAge, maxAge).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
