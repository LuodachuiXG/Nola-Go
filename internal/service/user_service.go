package service

import (
	"context"
	"errors"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"
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
//   - ip: 请求的 IP 地址
func (s *UserService) Login(
	ctx context.Context,
	username, password, ip string,
) (*response.AuthResponse, error) {
	// 查询用户
	user, err := s.userRepo.GetByUsername(ctx, username)

	if err != nil {
		return nil, err
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
		return nil, errors.New("令牌创建失败，请检查服务器日志")
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
