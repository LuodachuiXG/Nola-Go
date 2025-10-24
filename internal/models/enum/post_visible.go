package enum

// PostVisible 文章可见性
type PostVisible string

const (
	// PostVisibleVisible 可见
	PostVisibleVisible PostVisible = "VISIBLE"

	// PostVisibleHidden 隐藏
	PostVisibleHidden PostVisible = "HIDDEN"
)

func PostVisiblePtr(v PostVisible) *PostVisible {
	return &v
}
