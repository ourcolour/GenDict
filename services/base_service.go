package services

import (
	"context"
	"goDict/dao"
	"goDict/models"
)

// IBaseService 泛型Service接口
type IBaseService[T models.IEntity] interface {
	Insert(ctx context.Context, entity *T) error
	DeleteById(ctx context.Context, id uint) error
	DeleteByIdList(ctx context.Context, ids []uint) error
	Update(ctx context.Context, entity *T) error
	SelectById(ctx context.Context, id uint) (*T, error)
	SelectByIdList(ctx context.Context, ids []uint, queryOptions *models.QueryOption) ([]*T, error)
	SelectByQuery(ctx context.Context, query *T, queryOptions *models.QueryOption) ([]*T, error)
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
}

// BaseService 泛型Service实现
type BaseService[T models.IEntity] struct {
	dao dao.IBaseDao[T]
}

// NewBaseService 创建新的BaseService实例
func NewBaseService[T models.IEntity](dao dao.IBaseDao[T]) *BaseService[T] {
	return &BaseService[T]{
		dao: dao,
	}
}

// Insert 插入记录
func (s *BaseService[T]) Insert(ctx context.Context, entity *T) error {
	return s.dao.Insert(ctx, entity)
}

// DeleteById 根据ID删除
func (s *BaseService[T]) DeleteById(ctx context.Context, id uint) error {
	return s.dao.DeleteById(ctx, id)
}

// DeleteByIdList 根据ID列表批量删除
func (s *BaseService[T]) DeleteByIdList(ctx context.Context, ids []uint) error {
	return s.dao.DeleteByIdList(ctx, ids)
}

// Update 更新记录
func (s *BaseService[T]) Update(ctx context.Context, entity *T) error {
	return s.dao.Update(ctx, entity)
}

// SelectById 根据ID查询
func (s *BaseService[T]) SelectById(ctx context.Context, id uint) (*T, error) {
	return s.dao.SelectById(ctx, id)
}

// SelectByIdList 根据ID列表批量查询
func (s *BaseService[T]) SelectByIdList(ctx context.Context, ids []uint, queryOptions *models.QueryOption) (*models.QueryResult[T], error) {
	return s.dao.SelectByIdList(ctx, ids, queryOptions)
}

// SelectByQuery 根据条件查询
func (s *BaseService[T]) SelectByQuery(ctx context.Context, query *T, queryOptions *models.QueryOption) (*models.QueryResult[T], error) {
	return s.dao.SelectByQuery(ctx, query, queryOptions)
}

// WithTransaction 执行事务操作
func (s *BaseService[T]) WithTransaction(ctx context.Context, txFunc func(txCtx context.Context) error) error {
	return s.dao.WithTransaction(ctx, txFunc)
}
