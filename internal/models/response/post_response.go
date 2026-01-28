package response

import (
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/util"
)

// PostResponse 文章响应体
type PostResponse struct {
	PostId              uint             `json:"postId"`
	Title               string           `json:"title"`
	AutoGenerateExcerpt bool             `json:"autoGenerateExcerpt"`
	Excerpt             string           `json:"excerpt"`
	Slug                string           `json:"slug"`
	Cover               *string          `json:"cover"`
	AllowComment        bool             `json:"allowComment"`
	Pinned              *bool            `json:"pinned"`
	Status              enum.PostStatus  `json:"status"`
	Visible             enum.PostVisible `json:"visible"`
	Encrypted           bool             `json:"encrypted"`
	Password            *string          `json:"password"`
	Visit               uint             `json:"visit"`
	Category            *models.Category `json:"category"`
	Tags                []*models.Tag    `json:"tags"`
	CreateTime          int64            `json:"createTime"`
	LastModifyTime      *int64           `json:"lastModifyTime"`
}

// NewPostResponse 新建文章响应体，通过 *models.Post 文章
func NewPostResponse(post *models.Post) *PostResponse {
	isEncrypted := !util.StringIsNilOrBlank(post.Password)
	return &PostResponse{
		PostId:              post.PostId,
		Title:               post.Title,
		AutoGenerateExcerpt: post.AutoGenerateExcerpt,
		Excerpt:             post.Excerpt,
		Slug:                post.Slug,
		Cover:               post.Cover,
		AllowComment:        post.AllowComment,
		Pinned:              &post.Pinned,
		Status:              post.Status,
		Visible:             post.Visible,
		Encrypted:           isEncrypted,
		Password:            nil,
		Visit:               post.Visit,
		Category:            nil,
		Tags:                []*models.Tag{},
		CreateTime:          post.CreateTime,
		LastModifyTime:      post.LastModifyTime,
	}
}

// NewPostResponses 新建文章响应体，通过 []*models.Post
func NewPostResponses(posts []*models.Post) []*PostResponse {
	var responses []*PostResponse
	for _, post := range posts {
		responses = append(responses, NewPostResponse(post))
	}
	return responses
}

// NewPostMetaData 新建文章导出元数据
func NewPostMetaData(post PostResponse) models.PostMetaData {
	ret := models.PostMetaData{
		Title:               post.Title,
		AutoGenerateExcerpt: post.AutoGenerateExcerpt,
		Excerpt:             post.Excerpt,
		Slug:                post.Slug,
		Cover:               post.Cover,
		AllowComment:        post.AllowComment,
		Pinned:              *util.DefaultPtr(post.Pinned, false),
		Status:              post.Status,
		Visible:             post.Visible,
		Visit:               post.Visit,
		CreateTime:          post.CreateTime,
	}

	if post.Category != nil {
		ret.Category = &models.CategoryMetaData{
			DisplayName: post.Category.DisplayName,
			Slug:        post.Category.Slug,
		}
	}

	var tags []models.TagMetaData
	for _, tag := range post.Tags {
		if tag == nil {
			continue
		}
		tags = append(tags, models.TagMetaData{
			DisplayName: tag.DisplayName,
			Slug:        tag.Slug,
		})
	}

	ret.Tags = tags

	return ret
}
