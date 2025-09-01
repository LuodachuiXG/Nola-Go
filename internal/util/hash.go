package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SaltedHash 哈希盐值
type SaltedHash struct {
	Hash string
	Salt string
}

// GenerateSaltedHash 生成加盐哈希
//
//   - value: 待哈希的原始字符串数据
//   - saltBytes: 盐的字节长度
func GenerateSaltedHash(value string, saltBytes int) (*SaltedHash, error) {
	if saltBytes <= 0 {
		return nil, fmt.Errorf("saltBytes must be > 0")
	}
	// 生成盐值
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	// 转十六进制
	saltHex := hex.EncodeToString(salt)
	// 在数据前加上盐值
	hash := sha256.Sum256([]byte(saltHex + value))
	return &SaltedHash{
		Hash: hex.EncodeToString(hash[:]),
		Salt: saltHex,
	}, nil
}

// GenerateHash 生成加盐哈希
//   - value: 待哈希的原始字符串数据
func GenerateHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// VerifySaltedHash 验证哈希值是否匹配
//   - value: 待验证的原始字符串数据
//   - saltedHash: 哈西盐值
func VerifySaltedHash(value string, saltedHash *SaltedHash) bool {
	if saltedHash == nil {
		return false
	}

	hash := sha256.Sum256([]byte(saltedHash.Salt + value))
	return hex.EncodeToString(hash[:]) == saltedHash.Hash
}
