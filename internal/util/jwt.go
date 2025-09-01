package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMaker 提供简单的签发和解析方法（无状态）
// 此结构体仅提供简单的 JWT 签发和验证，没有 userId 与 Token 绑定验证
type JWTMaker struct {
	Secret        string
	Issuer        string
	Audience      string
	ExpireMinutes time.Duration
}

// Sign 为指定用户 ID 生成 Token
func (m JWTMaker) Sign(userId uint, username string, extra map[string]string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":      m.Issuer,
		"aud":      m.Audience,
		"iat":      now.Unix(),
		"exp":      now.Add(time.Minute * m.ExpireMinutes).Unix(),
		"user_id":  userId,
		"username": username,
	}

	for k, v := range extra {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.Secret))
}

// Parse 解析并验证 Token，成功返回 Claims
func (m JWTMaker) Parse(tokenStr string) (jwt.MapClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}
	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.Secret), nil
	}

	token, err := jwt.Parse(tokenStr, keyFunc)
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

	return claims, nil
}
