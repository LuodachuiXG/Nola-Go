package models

// Pair 键值对结构体
type Pair[T, U any] struct {
	First  T `json:"first"`
	Second U `json:"second"`
}
