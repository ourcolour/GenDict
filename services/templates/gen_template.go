//go:build exclude
// +build exclude
package services

import (
	"context"
	"goDict/dao"
	"goDict/models"
	"gorm.io/gorm"
)

// {{.UpperModelName}}Service {{.UpperModelName}}服务（可扩展自定义方法）
type {{.UpperModelName}}Service struct {
    BaseService[models.{{.UpperModelName}}]
}

// New{{.UpperModelName}}Service 创建{{.UpperModelName}}服务
func New{{.UpperModelName}}Service(db *gorm.DB) *{{.UpperModelName}}Service {
    {{.lowerModelName}}Dao := dao.NewBaseDao[models.{{.UpperModelName}}](db)
    baseService := NewBaseService[models.{{.UpperModelName}}]({{.lowerModelName}}Dao)
    return &{{.UpperModelName}}Service{BaseService: *baseService}
}

// CustomQueryByField 自定义查询方法示例
func (s *{{.UpperModelName}}Service) CustomQueryByField(ctx context.Context, fieldName string, value interface{}) ([]*models.{{.UpperModelName}}, error) {
	var results []*models.{{.UpperModelName}}
	db := dao.GetDbFromContext(ctx, s.dao.(*dao.BaseDao[models.{{.UpperModelName}}]).GetDb())

	result := db.WithContext(ctx).Table(s.dao.GetTableName()).Where(fieldName+" = ?", value).Find(&results)
	if result.Error != nil {
		return nil, result.Error
	}

	return results, nil
}