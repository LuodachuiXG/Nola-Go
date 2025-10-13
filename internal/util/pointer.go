package util

// StringPtr 返回一个指向字符串的指针
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr 返回一个指向 int64 的指针
func Int64Ptr(i int64) *int64 {
	return &i
}
