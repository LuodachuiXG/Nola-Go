package response

import (
	"nola-go/internal/models"
)

// PostApiResponse 博客前端文章响应体
type PostApiResponse struct {
	PostId         uint             `json:"postId"`
	Title          string           `json:"title"`
	Excerpt        *string          `json:"excerpt"`
	Slug           string           `json:"slug"`
	Cover          *string          `json:"cover"`
	AllowComment   bool             `json:"allowComment"`
	Pinned         *bool            `json:"pinned"`
	Encrypted      bool             `json:"encrypted"`
	Visit          uint             `json:"visit"`
	Category       *models.Category `json:"category"`
	Tags           []*models.Tag    `json:"tags"`
	CreateTime     int64            `json:"createTime"`
	LastModifyTime *int64           `json:"lastModifyTime"`
}

// NewPostApiResponse 新建博客前端文章响应体，通过 *response.PostResponse 文章响应体
//
// Parameters:
//   - post: 文章响应体
//   - hide: 是否隐藏敏感信息（如果文章有密码且此项设置 true，则隐藏。如果文章没有密码，此项设置无效）
func NewPostApiResponse(post *PostResponse, hide bool) *PostApiResponse {

	var excerpt = &post.Excerpt
	var lastModifyTime = post.LastModifyTime

	if post.Encrypted && hide {
		// 如果文章加密，不返回摘要
		excerpt = nil
		// 如果文章加密，不返回最后修改时间
		lastModifyTime = nil
	}

	return &PostApiResponse{
		PostId:         post.PostId,
		Title:          post.Title,
		Excerpt:        excerpt,
		Slug:           post.Slug,
		Cover:          post.Cover,
		AllowComment:   post.AllowComment,
		Pinned:         post.Pinned,
		Encrypted:      post.Encrypted,
		Visit:          post.Visit,
		Category:       post.Category,
		Tags:           post.Tags,
		CreateTime:     post.CreateTime,
		LastModifyTime: lastModifyTime,
	}
}

// NewPostApiResponses 新建博客前端文章响应体，通过 []*response.PostResponse
//
// Parameters:
//   - post: 文章响应体
//   - hide: 是否隐藏敏感信息（如果文章有密码且此项设置 true，则隐藏。如果文章没有密码，此项设置无效）
func NewPostApiResponses(posts []*PostResponse, hide bool) []*PostApiResponse {
	var responses []*PostApiResponse
	for _, post := range posts {
		responses = append(responses, NewPostApiResponse(post, hide))
	}
	return responses
}
