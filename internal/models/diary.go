package models

// Diary 日记
type Diary struct {
	// DiaryId 日记 ID
	DiaryId uint `gorm:"column:diary_id;primaryKey;autoIncrement" json:"diaryId"`
	// Content 日记内容
	Content string `gorm:"column:content;type:text COLLATE utf8mb4_general_ci;not null" json:"content"`
	// Html（由 content 解析得来）
	Html string `gorm:"column:html;type:text COLLATE utf8mb4_general_ci;not null" json:"html"`
	// CreateTime 创建时间
	CreateTime int64 `gorm:"column:create_time;autoCreateTime:milli;not null" json:"createTime"`
	// LastModifyTime 最后修改时间
	LastModifyTime *int64 `gorm:"column:last_modify_time" json:"lastModifyTime"`
}

func (Diary) TableName() string {
	return "diary"
}
