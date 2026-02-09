package util

// Map 重建数组
func Map[T, U any](ts []T, fn func(T) U) []U {
	result := make([]U, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

// Filter 过滤数组
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

// Find 找到第一个符合条件的元素
func Find[T any](slice []T, predicate func(T) bool) *T {
	for _, item := range slice {
		if predicate(item) {
			return &item
		}
	}

	return nil
}

// Chunk 切割数组
func Chunk[T any](slice []T, size int) [][]T {

	if size <= 0 {
		panic("size 必须大于 0")
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// AssociateBy 创建一个映射，将数组中的元素映射到指定的键
func AssociateBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T)
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = item
	}
	return result
}

// DefaultEmptySlice 默认空数组
// 如果 slice 为 nil，则返回一个空数组
func DefaultEmptySlice[T any](slice []T) []T {
	if slice == nil {
		return []T{}
	}
	return slice
}
