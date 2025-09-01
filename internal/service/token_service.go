package service

import (
	"errors"
	"nola-go/internal/config"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService 用于管理 Token 的签发与 userId -> token 的绑定验证
type TokenService struct {
	secret   string
	issuer   string
	audience string
	expires  time.Duration
	lock     sync.RWMutex
	tokenMap map[uint]string
}

// NewTokenService 创建 TokenService
func NewTokenService(config config.JWTConfig) *TokenService {
	return &TokenService{
		secret:   config.Secret,
		issuer:   config.Issuer,
		audience: config.Audience,
		expires:  config.ExpireMinutes,
		tokenMap: make(map[uint]string),
	}
}

// Generate 为指定用户签发 JWT，并把签发的 Token 与 userId 绑定
//   - userId: 用户 ID
//   - username: 用户名
//   - extra: 令牌额外附带信息
func (s *TokenService) Generate(userId uint, username string, extra map[string]string) (string, error) {
	now := time.Now()
	exp := now.Add(s.expires)

	claims := jwt.MapClaims{
		"aud":      s.audience,
		"iss":      s.issuer,
		"iat":      now.Unix(),
		"exp":      exp.Unix(),
		"user_id":  userId,
		"username": username,
	}
	for k, v := range extra {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", err
	}

	// 将 token 与 userId 绑定（同时使可能存在的旧 Token 过期）
	s.lock.Lock()
	s.tokenMap[userId] = ss
	s.lock.Unlock()

	return ss, nil
}

// Verify 验证用户 ID 与 Token 是否匹配
//   - userId: 用户 ID
//   - token: Token
func (s *TokenService) Verify(userId uint, token string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if token == "" {
		return false
	}
	stored, ok := s.tokenMap[userId]
	if !ok {
		return false
	}
	return stored == token
}

// ParseAndValidate 解析 Token 并验证 Audience 和 Issuer
// 解析成功返回 MapClaims，否则返回错误
//   - tokenString: Token 字符串
func (s *TokenService) ParseAndValidate(tokenString string) (jwt.MapClaims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	}

	token, err := jwt.Parse(tokenString, keyFunc, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// 检查 audience / issuer
	if aud, ok := claims["aud"]; ok {
		if audStr, ok := aud.(string); ok {
			if audStr != s.audience {
				return nil, errors.New("invalid audience")
			}
		}
	}

	if iss, ok := claims["iss"]; ok {
		if issStr, ok := iss.(string); ok {
			if issStr != s.issuer {
				return nil, errors.New("invalid issuer")
			}
		}
	}
	return claims, nil
}
