package enum

// PostStatus 文章状态
type PostStatus string

const (
	// PostStatusPublished 已发布
	PostStatusPublished PostStatus = "PUBLISHED"

	// PostStatusDraft 草稿
	PostStatusDraft PostStatus = "DRAFT"

	// PostStatusDeleted 已删除（回收站）
	PostStatusDeleted PostStatus = "DELETED"
)

func PostStatusPtr(s PostStatus) *PostStatus {
	return &s
}
