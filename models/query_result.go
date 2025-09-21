package models

import "math"

// Pagination 分页响应结构
type QueryResult[T IEntity] struct {
	Data       []*T  `json:"data"`       // 数据列表
	PageNum    int   `json:"pageNum"`    // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页大小
	TotalCount int64 `json:"totalCount"` // 总数
	PageCount  int   `json:"pageCount"`  // 总页数
}

func NewQueryResult[T IEntity](data []*T, totalCount int64, queryOptions *QueryOption) *QueryResult[T] {
	pageNum := queryOptions.GetPageNum()
	pageSize := queryOptions.GetPageSize()

	// 计算共计多少页
	pageCount := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &QueryResult[T]{
		Data:       data,
		PageNum:    pageNum,
		PageSize:   pageSize,
		TotalCount: totalCount,
		PageCount:  pageCount,
	}
}
