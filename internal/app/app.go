package app

import (
	"fmt"
	"nola-go/internal/config"
	"nola-go/internal/db"
	"nola-go/internal/logger"
	"nola-go/internal/middleware"
	"nola-go/internal/repository"
	"nola-go/internal/router"
	"nola-go/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Nola struct {
	Config          *config.Config
	DB              *gorm.DB
	Redis           *redis.Client
	UserRepo        repository.UserRepository
	PostRepo        repository.PostRepository
	ConfigRepo      repository.ConfigRepository
	TagRepo         repository.TagRepository
	CategoryRepo    repository.CategoryRepository
	LinkRepo        repository.LinkRepository
	TokenService    *service.TokenService
	UserService     *service.UserService
	PostService     *service.PostService
	ConfigService   *service.ConfigService
	TagService      *service.TagService
	CategoryService *service.CategoryService
	LinkService     *service.LinkService
	Engine          *gin.Engine
}

// NewNola 创建 Nola 实例
func NewNola() (*Nola, error) {
	a := &Nola{}

	// 初始化日志组件 Zap
	zap := logger.InitLogger()
	defer func() {
		_ = zap.Sync()
	}()

	// 读取配置文件
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("配置文件 config.yaml 读取失败: %w", err)
	}
	a.Config = cfg

	// 初始化 MySQL
	database, err := db.ConnectMySQL(cfg)
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
	a.ConfigRepo = repository.NewConfigRepository(a.DB)
	a.TagRepo = repository.NewTagRepository(a.DB)
	a.CategoryRepo = repository.NewCategoryRepository(a.DB)
	a.PostRepo = repository.NewPostRepository(a.DB, a.TagRepo, a.CategoryRepo)
	a.LinkRepo = repository.NewLinkRepository(a.DB)

	// Service
	a.TokenService = service.NewTokenService(a.Config.JWT)
	a.UserService = service.NewUserService(a.UserRepo, a.TokenService)
	a.ConfigService = service.NewConfigService(a.ConfigRepo)
	a.TagService = service.NewTagService(a.TagRepo)
	a.CategoryService = service.NewCategoryService(a.CategoryRepo)
	a.PostService = service.NewPostService(a.PostRepo, a.TagService, a.CategoryService)
	a.LinkService = service.NewLinkService(a.LinkRepo)

	r := gin.New()

	// 替换 Gin 默认日志组件
	r.Use(middleware.ZapLogger(zap), middleware.ZapRecovery(zap, true))

	// 设置路由
	router.SetupRouters(r, &router.Deps{
		TokenService:    a.TokenService,
		UserService:     a.UserService,
		PostService:     a.PostService,
		ConfigService:   a.ConfigService,
		TagService:      a.TagService,
		CategoryService: a.CategoryService,
		LinkService:     a.LinkService,
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
