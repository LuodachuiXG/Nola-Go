package repository

import (
	"context"
	"nola-go/internal/models"

	"gorm.io/gorm"
)

type PostRepository interface {
	List(ctx context.Context, limit, offset int) ([]models.Post, int64, error)
	Get(ctx context.Context, id uint) (*models.Post, error)
	Create(ctx context.Context, p *models.Post) error
	Update(ctx context.Context, p *models.Post) error
	Delete(ctx context.Context, id uint) error
}

type postRepo struct {
	db *gorm.DB
}

func (r *postRepo) List(ctx context.Context, limit, offset int) ([]models.Post, int64, error) {
	var ps []models.Post
	var total int64
	tx := r.db.WithContext(ctx).Model(&models.Post{})
	tx.Count(&total)
	if err := tx.Limit(limit).Offset(offset).Order("create_time desc").Find(&ps).Error; err != nil {
		return nil, 0, err
	}
	return ps, total, nil
}

func (r *postRepo) Get(ctx context.Context, id uint) (*models.Post, error) {
	var p models.Post
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postRepo) Create(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *postRepo) Update(ctx context.Context, p *models.Post) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *postRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(id).Error
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepo{db: db}
}
