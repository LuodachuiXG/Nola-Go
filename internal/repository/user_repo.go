package repository

import (
	"context"
	"errors"
	"nola-go/internal/models"
	"nola-go/internal/models/request"
	"nola-go/internal/util"

	"gorm.io/gorm"
)

// UserRepository 用户 Repo 接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, u *models.User) error
	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, userId uint, userInfo *request.UserInfoRequest) (bool, error)
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, userId uint, hash *util.SaltedHash) (bool, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	// GetById 根据用户 ID 获取用户
	GetById(ctx context.Context, userId uint) (*models.User, error)
	// GetAllUsers 获取所有用户
	GetAllUsers(ctx context.Context) ([]*models.User, error)
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository 创建用户 Repo
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// UpdateUser 更新用户信息
func (r *userRepo) UpdateUser(ctx context.Context, userId uint, userInfo *request.UserInfoRequest) (bool, error) {
	ret := r.db.WithContext(ctx).Model(&models.User{}).Where("user_id = ?", userId).Updates(models.User{
		Avatar:      userInfo.Avatar,
		Description: userInfo.Description,
		DisplayName: userInfo.DisplayName,
		Email:       userInfo.Email,
		Username:    userInfo.Username,
	})
	return ret.RowsAffected > 0, ret.Error
}

// UpdatePassword 更新用户密码
func (r *userRepo) UpdatePassword(ctx context.Context, userId uint, hash *util.SaltedHash) (bool, error) {
	ret := r.db.WithContext(ctx).Model(&models.User{}).Where("user_id = ?", userId).Updates(models.User{
		Password: hash.Hash,
		Salt:     hash.Salt,
	})
	return ret.RowsAffected > 0, ret.Error
}

// GetByUsername 根据用户名获取用户
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetById 根据用户 ID 获取用户
func (r *userRepo) GetById(ctx context.Context, userId uint) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetAllUsers 获取所有用户
func (r *userRepo) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
