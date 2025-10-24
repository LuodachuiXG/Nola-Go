package enum

// PostSort 文章排序
type PostSort string

const (
	// PostSortCreateDesc 创建时间降序
	PostSortCreateDesc PostSort = "CREATE_DESC"
	// PostSortCreateAsc 创建时间升序
	PostSortCreateAsc PostSort = "CREATE_ASC"
	// PostSortModifyDesc 修改时间降序
	PostSortModifyDesc PostSort = "MODIFY_DESC"
	// PostSortModifyAsc 修改时间升序
	PostSortModifyAsc PostSort = "MODIFY_ASC"
	// PostSortVisitDesc 访问量降序
	PostSortVisitDesc PostSort = "VISIT_DESC"
	// PostSortVisitAsc 访问量升序
	PostSortVisitAsc PostSort = "VISIT_ASC"
	// PostSortPinned 置顶排序
	PostSortPinned PostSort = "PINNED"
)

func PostSortPtr(s PostSort) *PostSort {
	return &s
}
