package app

import (
	"fmt"
	"log"
	"nola-go/internal/config"
	"nola-go/internal/db"
	"nola-go/internal/repository"
	"nola-go/internal/router"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Nola struct {
	Config       *config.Config
	DB           *gorm.DB
	Redis        *redis.Client
	UserRepo     repository.UserRepository
	PostRepo     repository.PostRepository
	TokenService *service.TokenService
	UserService  *service.UserService
	PostService  *service.PostService
	Engine       *gin.Engine
}

// NewNola 创建 Nola 实例
func NewNola() (*Nola, error) {
	a := &Nola{}

	// 读取配置文件
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("配置文件 config.yaml 读取失败: %w", err)
	}
	a.Config = cfg

	// 初始化 MySQL
	database, err := db.ConnectMySQL(cfg)
	log.Println(cfg.MySQL.DSN)
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}
	a.DB = database

	// 初始化 Redis
	redisClient, err := db.ConnectRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}
	a.Redis = redisClient

	// Repository
	a.UserRepo = repository.NewUserRepository(a.DB)
	a.PostRepo = repository.NewPostRepository(a.DB)

	// Service
	a.TokenService = service.NewTokenService(a.Config.JWT)
	a.UserService = service.NewUserService(a.UserRepo, a.TokenService)
	a.PostService = service.NewPostService(a.PostRepo)

	r := gin.Default()
	// 设置路由
	router.SetupRouters(r, &router.Deps{
		TokenService: a.TokenService,
		UserService:  a.UserService,
		PostService:  a.PostService,
	})

	// 只信任 本机代理
	err = r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return nil, fmt.Errorf("设置 Trusted Proxies 失败: %w", err)
	}

	a.Engine = r

	return a, nil
}

// Run 启动 Nola 服务器
func (n *Nola) Run() error {
	err := n.Engine.Run(n.Config.Server.Address())
	if err != nil {
		return fmt.Errorf("服务器启动失败: %w", err)
	}
	return nil
}
