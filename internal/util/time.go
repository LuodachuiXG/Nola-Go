package util

import "time"

// FormatDate 格式化日期为  2024-07-28 的形式
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
