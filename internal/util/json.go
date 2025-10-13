package util

import "encoding/json"

// ToJsonString 将任意类型转为 JSON 字符串
func ToJsonString(v any) *string {
	str, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return StringPtr(string(str))
}

// FromJsonString 将 JSON 字符串转为任意类型
func FromJsonString(jsonStr *string, v any) error {
	return json.Unmarshal([]byte(*jsonStr), v)
}
