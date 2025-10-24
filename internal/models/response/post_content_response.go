package response

import (
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
)

// PostContentResponse 文章内容响应数据类
type PostContentResponse struct {
	// PostContentId 文章内容 ID
	PostContentId uint `json:"postContentId"`
	// PostId 文章 ID
	PostId uint `json:"postId"`
	// Status 文章内容状态
	Status *enum.PostContentStatus `json:"status"`
	// DraftName 草稿名称
	DraftName *string `json:"draftName"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `json:"lastModifyTime"`
}

// NewPostContentResponse 创建文章内容响应数据类，通过 *models.PostContent
func NewPostContentResponse(postContent *models.PostContent) *PostContentResponse {
	return &PostContentResponse{
		PostContentId:  postContent.PostContentId,
		PostId:         postContent.PostId,
		Status:         enum.PostContentStatusPtr(postContent.Status),
		DraftName:      postContent.DraftName,
		LastModifyTime: postContent.LastModifyTime,
	}
}

// NewPostContentResponses 创建文章内容响应数据类，通过 []*models.PostContent
func NewPostContentResponses(postContents []*models.PostContent) []*PostContentResponse {
	res := make([]*PostContentResponse, len(postContents))
	for i, content := range postContents {
		res[i] = NewPostContentResponse(content)
	}
	return res
}
