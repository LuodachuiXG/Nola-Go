package models

// Post 文章
type Post struct {
	// PostId 文章 ID
	PostId uint `gorm:"column:post_id;primaryKey;autoIncrement" json:"postId"`

	// Title 标题
	Title string `gorm:"column:title;size:256;not null" json:"title"`

	// AutoGenerateExcerpt 是否自动生成摘要
	AutoGenerateExcerpt bool `gorm:"column:auto_generate_excerpt;not null" json:"autoGenerateExcerpt"`

	// Excerpt 摘要
	Excerpt string `gorm:"column:excerpt;size:1024" json:"excerpt"`

	// Slug 别名
	Slug string `gorm:"column:slug;size:128;uniqueIndex;not null" json:"slug"`

	// Cover 封面
	Cover *string `gorm:"column:cover;size:512" json:"cover,omitempty"`

	// AllowComment 是否允许评论
	AllowComment bool `gorm:"column:allow_comment;not null" json:"allowComment"`

	// Pinned 是否置顶
	Pinned bool `gorm:"column:pinned;not null" json:"pinned"`

	// Status 状态
	Status PostStatus `gorm:"column:status;type:varchar(24);not null" json:"status"`

	// Visible 可见性
	Visible PostVisible `gorm:"column:visible;type:varchar(24);not null" json:"visible"`

	// Password 密码
	Password *string `gorm:"column:password;size:64" json:"password,omitempty"`

	// Visit 访问量
	Visit uint `gorm:"column:visit;default:0;not null" json:"visit"`

	// CreateTime 创建时间
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`

	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (Post) TableName() string {
	return "post"
}

// PostStatus 文章状态
type PostStatus string

const (
	// PostPublished 已发布
	PostPublished PostStatus = "PUBLISHED"

	// PostDraft 草稿
	PostDraft PostStatus = "DRAFT"

	// PostDeleted 已删除（回收站）
	PostDeleted PostStatus = "DELETED"
)

// PostVisible 文章可见性
type PostVisible string

const (
	// Visible 可见
	Visible PostVisible = "VISIBLE"

	// Hidden 隐藏
	Hidden PostVisible = "HIDDEN"
)
