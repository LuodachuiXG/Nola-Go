package util

import "regexp"

// StringPtr 返回一个指向字符串的指针
func StringPtr(s string) *string {
	return &s
}

// StringIsEmail 判断一个字符串是否为邮箱
func StringIsEmail(email string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email)
}

// StringIsNumberAndChar 判断一个字符串是否只由数字和字符组成
func StringIsNumberAndChar(s string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s)
}
