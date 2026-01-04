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

// DefaultEmptySlice 默认空数组
// 如果 slice 为 nil，则返回一个空数组
func DefaultEmptySlice[T any](slice []T) []T {
	if slice == nil {
		return []T{}
	}
	return slice
}
