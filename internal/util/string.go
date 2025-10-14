package util

import (
	"regexp"
	"strings"
)

// StringDefault 返回字符串或者默认值
// 如果字符串为 nil，则返回默认值
func StringDefault(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}

// StringIsEmail 判断一个字符串是否为邮箱
func StringIsEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// StringIsNumberAndChar 判断一个字符串是否只由数字和字符组成
func StringIsNumberAndChar(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}

// StringIsNilOrBlank 判断一个字符串是否为 nil 或者空白字符串
func StringIsNilOrBlank(s *string) bool {
	return s == nil || len(*s) == 0 || StringIsBlank(*s)
}

// StringIsBlank 判断一个字符串是否为空白字符串
func StringIsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
