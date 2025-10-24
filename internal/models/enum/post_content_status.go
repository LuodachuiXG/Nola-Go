package enum

// PostContentStatus 文章内容状态
type PostContentStatus string

const (
	// PostContentStatusPublished 文章内容已发布
	PostContentStatusPublished PostContentStatus = "PUBLISHED"
	// PostContentStatusDraft 文章内容草稿
	PostContentStatusDraft PostContentStatus = "DRAFT"
)

func PostContentStatusPtr(s PostContentStatus) *PostContentStatus {
	return &s
}
