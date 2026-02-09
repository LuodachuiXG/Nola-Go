package models

// Comment 评论
type Comment struct {
	// CommentId 评论 ID
	CommentId uint `gorm:"column:comment_id;primaryKey;autoIncrement" json:"commentId"`
	// PostId 文章 ID
	PostId uint `gorm:"column:post_id;not null" json:"postId"`
	// ParentCommentId 父评论 ID
	ParentCommentId *uint `gorm:"column:parent_comment_id" json:"parentCommentId"`
	// ReplyCommentId 回复评论 ID
	ReplyCommentId *uint `gorm:"column:reply_comment_id" json:"replyCommentId"`
	// ReplyDisplayName 回复用户名
	ReplyDisplayName *string `gorm:"column:reply_display_name;size:128" json:"replyDisplayName"`
	// Content 评论内容
	Content string `gorm:"column:content;type:text;not null" json:"content"`
	// Site 评论人站点
	Site *string `gorm:"column:site;size:512" json:"site"`
	// Display 评论人名称
	DisplayName string `gorm:"column:display_name;size:128;not null" json:"displayName"`
	// Email 评论人邮箱
	Email string `gorm:"column:email;size:128;not null" json:"email"`
	// CreateTime 评论时间
	CreateTime int64 `gorm:"column:create_time;not null" json:"createTime"`
	// IsPass 是否通过审核
	IsPass bool `gorm:"column:is_pass;not null" json:"isPass"`
	// Children 子评论
	Children []Comment `gorm:"-" json:"children"`
	// PostTitle 文章标题
	PostTitle *string `gorm:"-" json:"postTitle"`
}

func (Comment) TableName() string {
	return "comment"
}
