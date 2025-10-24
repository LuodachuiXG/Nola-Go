package request

import (
	"nola-go/internal/models/enum"
	"nola-go/internal/util"
	"strings"
)

// PostRequest 文章请求结构体
type PostRequest struct {
	// PostId 文章 ID
	PostId *uint `json:"postId"`
	// Title 标题
	Title string `json:"title" binding:"required"`
	// AutoGenerateExcerpt 是否自动生成摘要
	AutoGenerateExcerpt bool `json:"autoGenerateExcerpt" binding:"required"`
	// Excerpt 摘要
	Excerpt *string `json:"excerpt"`
	// Slug 别名
	Slug string `json:"slug" binding:"required"`
	// AllowComment 是否允许评论
	AllowComment bool `json:"allowComment" binding:"required"`
	// Status 文章状态
	Status enum.PostStatus `json:"status" binding:"required"`
	// Visible 文章可见性
	Visible enum.PostVisible `json:"visible" binding:"required"`
	// Content 文章内容（Markdown 或普通文本）
	Content *string `json:"content"`
	// CategoryId 分类 ID
	CategoryId *uint `json:"categoryId"`
	// TagIds 标签 ID 数组
	TagIds []uint `json:"tagIds"`
	// Cover 封面
	Cover *string `json:"cover"`
	// Pinned 是否置顶
	Pinned bool `json:"pinned"`
	// Encrypted 文章是否加密（为 true 时需提供 password，为 nil 保持不变，为 false 删除密码）
	Encrypted *bool `json:"encrypted"`
	// Password 文章密码
	Password *string `json:"password"`
}

// NewPostRequestByNameAndContent 通过名称和内容创建文章请求体
//
// Parameters:
//   - name: 文章名称
//   - content: 文章内容
func NewPostRequestByNameAndContent(name string, content string) *PostRequest {

	// 防止文章名称为 Markdown 文件名，剔除掉后缀名
	var title string
	if strings.Contains(name, ".") {
		// 当前名称可能是文件名，剔除掉后缀
		title = strings.Split(name, ".")[0]
	} else {
		title = name
	}

	return &PostRequest{
		PostId:              nil,
		Title:               title,
		AutoGenerateExcerpt: true,
		Slug:                util.StringPostNameToSlug(name),
		Excerpt:             nil,
		AllowComment:        true,
		Status:              enum.PostStatusPublished,
		Visible:             enum.PostVisibleVisible,
		Content:             util.StringPtr(content),
		CategoryId:          nil,
		TagIds:              nil,
		Cover:               nil,
		Encrypted:           nil,
		Password:            nil,
	}
}
