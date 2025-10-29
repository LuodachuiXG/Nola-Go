package util

import (
	"math/rand"
	"regexp"
	"strings"

	"github.com/mozillazg/go-pinyin"
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

// StringIsNumber 判断一个字符串是否是数字
func StringIsNumber(s string) bool {
	return regexp.MustCompile(`^\d+$`).MatchString(s)
}

// StringIsNilOrBlank 判断一个字符串是否为 nil 或者空白字符串
func StringIsNilOrBlank(s *string) bool {
	return s == nil || len(*s) == 0 || StringIsBlank(*s)
}

// StringIsBlank 判断一个字符串是否为空白字符串
func StringIsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// StringChineseToPinyin 将汉字转为拼音
//
// Parameters:
//   - s: 待转换的汉字字符串
//   - separator: 分隔符（默认空）
func StringChineseToPinyin(s string, separator *string) string {
	// 去掉文本中的所有非数字、非字母、非汉字文本
	str := regexp.MustCompile(`[^0-9a-zA-Z\u4e00-\u9fa5]`).ReplaceAllString(s, "")

	// 默认模式（不带声调）
	a := pinyin.NewArgs()
	py := pinyin.Pinyin(str, a)

	var ret string
	for _, v := range py {
		ret += v[0] + StringDefault(separator, "")
	}
	return ret
}

// StringPostNameToSlug 将文章名称转为别名
func StringPostNameToSlug(name string) string {
	// 将中文转拼音
	py := StringChineseToPinyin(name, nil)
	// 将所有空白字符替换成 -，并且转小写
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.ToLower(py), "-")
}

// StringRandom 生成指定长度的随机字符，包括数字和字母
func StringRandom(length int) string {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	var ret string
	for i := 0; i < length; i++ {
		ret += string(chars[rand.Intn(len(chars))])
	}
	return ret
}
