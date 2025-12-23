package service

import (
	"context"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"

	"go.uber.org/zap"
)

type DiaryService struct {
	diaryRepo repository.DiaryRepository
}

// NewDiaryService 创建日记 Service
func NewDiaryService(diaryRepo repository.DiaryRepository) *DiaryService {
	return &DiaryService{
		diaryRepo: diaryRepo,
	}
}

// AddDiary 添加日记
func (s *DiaryService) AddDiary(c context.Context, diary *request.DiaryRequest) (*models.Diary, error) {
	ret, err := s.diaryRepo.AddDiary(c, diary)
	if err != nil {
		logger.Log.Error("添加日记失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// DeleteDiaries 删除日记
func (s *DiaryService) DeleteDiaries(c context.Context, diaryIds []uint) (bool, error) {
	ret, err := s.diaryRepo.DeleteDiaries(c, diaryIds)
	if err != nil {
		logger.Log.Error("删除日记失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdateDiary 更新日记
func (s *DiaryService) UpdateDiary(c context.Context, diary *request.DiaryRequest) (bool, error) {
	ret, err := s.diaryRepo.UpdateDiary(c, diary)
	if err != nil {
		logger.Log.Error("更新日记失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// Diaries 获取所有日记
func (s *DiaryService) Diaries(c context.Context, sort *enum.DiarySort) ([]*models.Diary, error) {
	ret, err := s.diaryRepo.Diaries(c, sort)
	if err != nil {
		logger.Log.Error("获取所有日记失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// DiariesPager 分页获取所有日记
func (s *DiaryService) DiariesPager(c context.Context, page, size int, sort *enum.DiarySort) (*models.Pager[models.Diary], error) {
	if page == 0 {
		// 获取所有日记
		diaries, err := s.Diaries(c, sort)
		if err != nil {
			return nil, err
		}

		return &models.Pager[models.Diary]{
			Page:       0,
			Size:       0,
			Data:       diaries,
			TotalData:  int64(len(diaries)),
			TotalPages: 1,
		}, nil
	}

	ret, err := s.diaryRepo.DiariesPager(c, page, size, sort)
	if err != nil {
		logger.Log.Error("分页获取所有日记失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// DiaryCount 获取日记数量
func (s *DiaryService) DiaryCount(c context.Context) (int64, error) {
	count, err := s.diaryRepo.DiaryCount(c)
	if err != nil {
		logger.Log.Error("获取日记数量失败", zap.Error(err))
		return 0, response.ServerError
	}
	return count, nil
}
