package dao

import (
	"context"
	"fmt"
	"goDict/models"
	"goDict/utils"
	"gorm.io/gorm"
	"log/slog"
	"reflect"
	"time"
)

// IBaseDao 泛型DAO接口
type IBaseDao[T models.IEntity] interface {
	Insert(ctx context.Context, entity *T) error
	DeleteById(ctx context.Context, id uint) error
	DeleteByIdList(ctx context.Context, ids []uint) error
	Update(ctx context.Context, entity *T) error
	SelectById(ctx context.Context, id uint) (*T, error)
	SelectByIdList(ctx context.Context, ids []uint, queryOptions *models.QueryOption) (*models.QueryResult[T], error)
	SelectByQuery(ctx context.Context, query *T, queryOptions *models.QueryOption) (*models.QueryResult[T], error)
	WithTransaction(ctx context.Context, txFunc func(txCtx context.Context) error) error
	GetTableName() string
	GetDb() *gorm.DB
}

// BaseDao 泛型DAO实现
type BaseDao[T models.IEntity] struct {
	db        *gorm.DB
	tableName string
}

// 确保 BaseDao 实现所有 IBaseDao 方法
var _ IBaseDao[models.IEntity] = (*BaseDao[models.IEntity])(nil)

// NewBaseDao 创建新的BaseDao实例
func NewBaseDao[T models.IEntity](db *gorm.DB) *BaseDao[T] {
	// 使用反射获取表名
	var entity T
	tableName := GetTableName(entity)

	return &BaseDao[T]{
		db:        db,
		tableName: tableName,
	}
}

// GetTableName 获取表名
func (d *BaseDao[T]) GetTableName() string {
	return d.tableName
}

// 辅助函数：通过反射获取表名
func GetTableName(entity interface{}) string {
	// 如果实体实现了 TableName 方法，则调用它
	if tableNamer, ok := entity.(interface{ TableName() string }); ok {
		return tableNamer.TableName()
	}

	// 否则，尝试通过反射获取表名
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 获取类型名称作为表名（小写并复数化）
	typeName := val.Type().Name()
	if len(typeName) > 0 {
		// 简单的复数化规则（实际应用中可能需要更复杂的规则）
		return fmt.Sprintf("%ss", utils.ToSnakeCase(typeName))
	}

	return "" // 默认表名
}

// 在 BaseDao 结构体下添加这个方法
func (d *BaseDao[T]) GetDb() *gorm.DB {
	return d.db
}

// Insert 插入记录
func (d *BaseDao[T]) Insert(ctx context.Context, entity *T) error {
	result := d.db.WithContext(ctx).Create(entity)
	return result.Error
}

// DeleteById 根据ID删除
func (d *BaseDao[T]) DeleteById(ctx context.Context, id uint) error {
	var entity T
	result := d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity)
	return result.Error
}

// DeleteByIdList 根据ID列表批量删除
func (d *BaseDao[T]) DeleteByIdList(ctx context.Context, ids []uint) error {
	var entity T
	result := d.db.WithContext(ctx).Where("id IN ?", ids).Delete(&entity)
	return result.Error
}

// Update 更新记录
func (d *BaseDao[T]) Update(ctx context.Context, entity *T) error {
	result := d.db.WithContext(ctx).Save(entity)
	return result.Error
}

// SelectById 根据ID查询
func (d *BaseDao[T]) SelectById(ctx context.Context, id uint) (*T, error) {
	var entity T
	result := d.db.WithContext(ctx).First(&entity, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &entity, nil
}

// SelectByIdList 根据ID列表批量查询
func (d *BaseDao[T]) SelectByIdList(ctx context.Context, ids []uint, queryOptions *models.QueryOption) (*models.QueryResult[T], error) {
	var entities []*T

	// 构建基础查询
	db := d.db.WithContext(ctx).Table(d.tableName).Where("id IN ?", ids)

	// 处理排序
	if nil != queryOptions && 0 < len(queryOptions.Sorting) {
		for _, sorting := range queryOptions.Sorting {
			db = db.Order(sorting)
		}
	}

	// 处理分页
	if nil != queryOptions {
		if queryOptions.PageSize != nil {
			db = db.Limit(*queryOptions.PageSize)
		}
		if queryOptions.PageNum != nil {
			db = db.Offset(queryOptions.GetOffset())
		}
	}

	// 执行查询获取数据
	result := db.Find(&entities)
	if nil != result.Error {
		return nil, result.Error
	}

	// 获取总记录数
	var totalCount int64
	countDb := d.db.WithContext(ctx).Table(d.tableName).Where("id IN ?", ids)
	countResult := countDb.Count(&totalCount)
	if nil != countResult.Error {
		return nil, countResult.Error
	}

	// 生成查询结果
	queryResult := models.NewQueryResult(entities, totalCount, queryOptions)

	return queryResult, nil
}

// SelectByQuery 根据条件查询
func (d *BaseDao[T]) SelectByQuery(ctx context.Context, query *T, queryOptions *models.QueryOption) (*models.QueryResult[T], error) {
	// 如果ctx为nil，使用Background作为默认上下文
	if ctx == nil {
		ctx = context.Background()
	}
	// 设置默认超时（如果上下文中没有超时）
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// 创建查询结果变量
	var entities []*T

	// 构建基础查询
	db := d.db.WithContext(ctx).Table(d.tableName)

	// 查询条件
	if nil != query {
		db = db.Where(query)
	}

	// 处理排序
	if nil != queryOptions && 0 < len(queryOptions.Sorting) {
		for _, sorting := range queryOptions.Sorting {
			db = db.Order(sorting)
		}
	}

	// 处理分页
	if nil != queryOptions {
		if queryOptions.PageSize != nil {
			db = db.Limit(*queryOptions.PageSize)
		}
		if queryOptions.PageNum != nil {
			db = db.Offset(queryOptions.GetOffset())
		}
	}

	// 执行查询获取数据
	result := db.Find(&entities)

	// 输出SQL语句
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&entities)
	})
	slog.Debug("SQL语句", "sql", sql)

	if nil != result.Error {
		return nil, result.Error
	}

	// 获取总记录数
	var totalCount int64
	countDb := d.db.WithContext(ctx).Table(d.tableName).Where(query)
	countResult := countDb.Count(&totalCount)
	if nil != countResult.Error {
		return nil, countResult.Error
	}

	// 生成查询结果
	queryResult := models.NewQueryResult(entities, totalCount, queryOptions)

	return queryResult, nil
}

// WithTransaction 事务支持
func (d *BaseDao[T]) WithTransaction(ctx context.Context, txFunc func(txCtx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return txFunc(txCtx)
	})
}

// GetDbFromContext 从上下文中获取数据库实例（支持事务）
func GetDbFromContext(ctx context.Context, defaultDb *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return defaultDb
}
