package models

const (
	DEFAULT_PAGE_NO   = 1
	DEFAULT_PAGE_SIZE = 10
)

type QueryOption struct {
	PageNum  *int
	PageSize *int
	Sorting  []string
}

func NewQueryOption() *QueryOption {
	pageNum := DEFAULT_PAGE_NO
	pageSize := DEFAULT_PAGE_SIZE

	return &QueryOption{
		PageNum:  &pageNum,
		PageSize: &pageSize,
		Sorting:  nil,
	}
}

func (q *QueryOption) SetPageNum(pageNum int) *QueryOption {
	q.PageNum = &pageNum
	return q
}

func (q *QueryOption) SetPageSize(pageSize int) *QueryOption {
	q.PageSize = &pageSize
	return q
}

func (q *QueryOption) SetSorting(sorting []string) *QueryOption {
	q.Sorting = sorting
	return q
}

func (q *QueryOption) GetOffset() int {
	// 页码
	pageNum := DEFAULT_PAGE_NO
	if nil != q.PageNum {
		pageNum = *q.PageNum
	}
	// 分页大小
	pageSize := DEFAULT_PAGE_SIZE
	if nil != q.PageSize {
		pageSize = *q.PageSize
	}

	return (pageNum - 1) * pageSize
}

func (q *QueryOption) GetPageNum() int {
	if q.PageNum == nil {
		return DEFAULT_PAGE_NO
	}
	return *q.PageNum
}

func (q *QueryOption) GetPageSize() int {
	if q.PageSize == nil {
		return DEFAULT_PAGE_SIZE
	}
	return *q.PageSize
}

func (q *QueryOption) GetSortingList() []string {
	if q.Sorting == nil {
		return []string{}
	}
	return q.Sorting
}
