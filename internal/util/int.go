package util

// CoerceAtLeast 确保 value 不小于 min
func CoerceAtLeast(value, min int) int {
	if value < min {
		return min
	}
	return value
}
