package request

// DiaryRequest 日记请求
type DiaryRequest struct {
	// DiaryId 日记 ID
	DiaryId *uint `json:"diaryId"`
	// Content 日记内容
	Content string `json:"content" binding:"required"`
}
