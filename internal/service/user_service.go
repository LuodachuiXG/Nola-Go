package service

import (
	"context"
	"errors"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"

	"go.uber.org/zap"
)

type UserService struct {
	userRepo     repository.UserRepository
	tokenService *TokenService
}

func NewUserService(userRepo repository.UserRepository, tokenService *TokenService) *UserService {
	return &UserService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

// Login 用户登录
//   - ctx: 上下文
//   - username: 用户名
//   - password: 密码
func (s *UserService) Login(
	ctx context.Context,
	username, password string,
) (*response.AuthResponse, error) {

	// 查询用户
	user, err := s.UserByUsername(ctx, username)

	if err != nil {
		return nil, errors.New("非法用户名或密码")
	}

	if user == nil {
		return nil, errors.New("非法用户名或密码")
	}

	// 验证密码合法性
	if !util.VerifySaltedHash(password, &util.SaltedHash{
		Hash: user.Password,
		Salt: user.Salt,
	}) {
		return nil, errors.New("非法用户名或密码")
	}

	// 生成 Token
	token, err := s.tokenService.Generate(user.UserId, user.Username, nil)
	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, response.ServerError
	}

	// 封装登录响应数据类
	return &response.AuthResponse{
		Username:      user.Username,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Description:   user.Description,
		CreateDate:    user.CreateDate,
		LastLoginDate: user.LastLoginDate,
		Avatar:        user.Avatar,
		Token:         token,
	}, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userId uint, userInfo *request.UserInfoRequest) (bool, error) {
	if !util.StringIsEmail(userInfo.Email) {
		return false, errors.New("邮箱格式错误")
	}
	if !util.StringIsNumberAndChar(userInfo.Username) {
		return false, errors.New("用户名只支持英文和数字")
	}
	if len(userInfo.Username) < 4 {
		return false, errors.New("用户名不能小于 4 位")
	}

	ret, err := s.userRepo.UpdateUser(ctx, userId, userInfo)
	if err != nil {
		logger.Logger.Error(err.Error())
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePassword 修改用户密码
func (s *UserService) UpdatePassword(ctx context.Context, userId uint, password string) (bool, error) {
	if len(password) < 8 {
		return false, errors.New("密码长度不能小于 8 位")
	}

	// 判断用户是否存在
	user, err := s.UserById(ctx, userId)
	if user == nil || err != nil {
		return false, errors.New("用户不存在")
	}

	// 对密码生成加盐哈希
	hash, err := util.GenerateSaltedHash(password, 32)

	if err != nil {
		logger.Logger.Error("密码生成失败", zap.Error(err))
		return false, response.ServerError
	}

	ret, err := s.userRepo.UpdatePassword(ctx, userId, hash)
	if err != nil {
		logger.Logger.Error(err.Error())
		return false, response.ServerError
	}

	return ret, nil
}

// UserById 根据用户 ID 获取用户
func (s *UserService) UserById(ctx context.Context, userId uint) (*models.User, error) {
	user, err := s.userRepo.GetById(ctx, userId)
	return user, err
}

// UserByUsername 根据用户名获取用户
func (s *UserService) UserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	return user, err
}

// AllUsers 获取所有用户
func (s *UserService) AllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepo.GetAllUsers(ctx)

	if err != nil {
		logger.Logger.Error(err.Error())
		return nil, response.ServerError
	}

	return users, nil
}
