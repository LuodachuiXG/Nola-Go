package util

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
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

// StringFileNameAddRandomSuffix 给文件名后，文件扩展名前加上 5 个随机字符，
// 用于在文件名已经存在时，防止文件名重复，
// 如果该文件名后已经有随机字符，则重新修改该随机字符为新的随机字符。
func StringFileNameAddRandomSuffix(s string) string {
	// 先获取文件的扩展名
	lastDotIndex := strings.LastIndex(s, ".")
	fileExt := ""
	fileName := s
	randomStr := StringRandom(5)

	if lastDotIndex != -1 {
		// 不包含点号
		fileExt = s[lastDotIndex+1:]
		fileName = s[:lastDotIndex]
	}

	// 文件名中是否之前已经加了随机字符
	isExistRandom := false

	if len(fileName) >= 6 {
		ret, err := regexp.MatchString("^_[a-z0-9]+$", fileName[len(fileName)-6:])
		if err != nil {
			isExistRandom = false
		} else {
			isExistRandom = ret
		}
	}

	if isExistRandom {
		// 文件名中已经存在随机字符，重新修改随机字符
		return fmt.Sprintf("%s_%s.%s", fileName[:len(fileName)-6], randomStr, fileExt)
	}

	// 文件名中不存在随机字符，直接在文件名后加上
	return fmt.Sprintf("%s_%s.%s", fileName, randomStr, fileExt)
}

// StringReplaceDoubleSlash 将所有双正反斜杠替换为当前系统的反斜杠
func StringReplaceDoubleSlash(path string) string {
	systemSlash := string(filepath.Separator)
	return strings.ReplaceAll(
		strings.ReplaceAll(path, "//", systemSlash),
		"\\\\", systemSlash,
	)
}

// StringFormatSlash 将所有反斜杠替换为当前系统的单斜杠
func StringFormatSlash(path string) string {
	systemSlash := string(filepath.Separator)
	result := strings.ReplaceAll(path, "\\", systemSlash)
	return StringReplaceDoubleSlash(result)
}

// StringSubstringAfterLast 获取字符串从指定字符后开始到字符串末尾的字符串
//   - str: 待处理的字符串
//   - delimiter: 分隔符
//
// Returns: 截取后的字符串
func StringSubstringAfterLast(str, delimiter string) string {
	index := strings.LastIndex(str, delimiter)
	if index == -1 {
		// 没有找到分隔符，返回原字符串
		return str
	}
	return str[index+len(delimiter):]
}

// StringRemoveSlash 删除所有正斜杠和反斜杠
func StringRemoveSlash(path string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(path, "/", ""),
		"\\", "",
	)
}

// StringIsUrl 判断文本是否是合法的 URL
func StringIsUrl(text string) bool {
	return regexp.MustCompile(`^(http|https)://[^.]+\.[^/]*$`).MatchString(text)
}

// StringToUint 将字符串转换为 uint
func StringToUint(s string) (uint, error) {
	var ret *uint
	ui, err := strconv.ParseUint(s, 10, 32)

	if err == nil {
		ret = new(uint)
		*ret = uint(ui)
		return *ret, nil
	}

	return 0, err
}
