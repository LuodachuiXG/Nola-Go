package repository

import (
	"context"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/util"
	"time"

	"gorm.io/gorm"
)

// DiaryRepository 日记 Repo 接口
type DiaryRepository interface {
	// AddDiary 添加日记
	AddDiary(ctx context.Context, diary *request.DiaryRequest) (*models.Diary, error)
	// DeleteDiaries 删除日记
	DeleteDiaries(ctx context.Context, diaryIds []uint) (bool, error)
	// UpdateDiary 更新日记
	UpdateDiary(ctx context.Context, diary *request.DiaryRequest) (bool, error)
	// Diaries 获取所有日记
	Diaries(ctx context.Context, sort *enum.DiarySort) ([]*models.Diary, error)
	// DiariesPager 分页获取所有日记
	DiariesPager(ctx context.Context, page, size int, sort *enum.DiarySort) (*models.Pager[models.Diary], error)
	// DiaryCount 日记数量
	DiaryCount(ctx context.Context) (int64, error)
}

type diaryRepo struct {
	db *gorm.DB
}

func NewDiaryRepository(db *gorm.DB) DiaryRepository {
	return &diaryRepo{
		db: db,
	}
}

// AddDiary 添加日记
func (r *diaryRepo) AddDiary(ctx context.Context, diary *request.DiaryRequest) (*models.Diary, error) {
	add := &models.Diary{
		Content:        diary.Content,
		Html:           util.MarkdownToHtml(diary.Content),
		CreateTime:     time.Now().UnixMilli(),
		LastModifyTime: nil,
	}

	err := r.db.WithContext(ctx).Create(add).Error
	if err != nil {
		return nil, err
	}

	return add, nil
}

// DeleteDiaries 删除日记
func (r *diaryRepo) DeleteDiaries(ctx context.Context, diaryIds []uint) (bool, error) {
	ret := r.db.WithContext(ctx).Where("diary_id IN ?", diaryIds).Delete(&models.Diary{})
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// UpdateDiary 更新日记
func (r *diaryRepo) UpdateDiary(ctx context.Context, diary *request.DiaryRequest) (bool, error) {
	updates := map[string]any{
		"content":          diary.Content,
		"html":             util.MarkdownToHtml(diary.Content),
		"last_modify_time": time.Now().UnixMilli(),
	}

	ret := r.db.WithContext(ctx).Model(&models.Diary{}).Where("diary_id = ?", diary.DiaryId).Updates(updates)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// Diaries 获取所有日记
func (r *diaryRepo) Diaries(ctx context.Context, sort *enum.DiarySort) ([]*models.Diary, error) {
	var diaries []*models.Diary
	err := r.sqlQueryDiaries(ctx, sort).Find(&diaries).Error
	if err != nil {
		return nil, err
	}
	return diaries, nil
}

// DiariesPager 分页获取所有日记
func (r *diaryRepo) DiariesPager(ctx context.Context, page, size int, sort *enum.DiarySort) (*models.Pager[models.Diary], error) {
	if page == 0 {
		// 获取所有日记
		data, err := r.Diaries(ctx, sort)
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.Diary]{
			Page:       0,
			Size:       0,
			Data:       data,
			TotalData:  int64(len(data)),
			TotalPages: 1,
		}, nil
	}

	pager, err := db.PagerBuilder[models.Diary](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return r.sqlQueryDiaries(ctx, sort)
	})

	if err != nil {
		return nil, err
	}

	return pager, nil
}

// DiaryCount 日记数量
func (r *diaryRepo) DiaryCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Diary{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// sqlQueryDiaries 构建连接查询 SQL
func (r *diaryRepo) sqlQueryDiaries(ctx context.Context, sort *enum.DiarySort) *gorm.DB {
	query := r.db.WithContext(ctx).Table("diary d")
	if sort != nil {
		switch *sort {
		case enum.DiarySortCreateTimeDesc:
			query = query.Order("d.create_time DESC")
		case enum.DiarySortCreateTimeAsc:
			query = query.Order("d.create_time ASC")
		case enum.DiarySortModifyTimeDesc:
			query = query.Order("d.last_modify_time DESC")
		case enum.DiarySortModifyTimeAsc:
			query = query.Order("d.last_modify_time ASC")
		}
	} else {
		// 默认创建时间降序
		query = query.Order("d.create_time DESC")
	}

	return query
}
