package service

import (
	"context"
	"nola-go/internal/models"
	"nola-go/internal/repository"
)

type PostService struct {
	posts repository.PostRepository
}

func NewPostService(p repository.PostRepository) *PostService {
	return &PostService{posts: p}
}

func (s *PostService) List(ctx context.Context, limit, offset int) ([]models.Post, int64, error) {
	return s.posts.List(ctx, limit, offset)
}

func (s *PostService) Get(ctx context.Context, id uint) (*models.Post, error) {
	return s.posts.Get(ctx, id)
}

func (s *PostService) Create(ctx context.Context, p *models.Post) error {
	return s.posts.Create(ctx, p)
}

func (s *PostService) Update(ctx context.Context, p *models.Post) error {
	return s.posts.Update(ctx, p)
}

func (s *PostService) Delete(ctx context.Context, id uint) error {
	return s.posts.Delete(ctx, id)
}
