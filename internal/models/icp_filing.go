package models

// ICPFiling ICP 备案信息
type ICPFiling struct {
	// ICP ICP 备案号
	ICP *string `json:"icp"`
	// Police 公网安备号
	Police *string `json:"police"`
	// Public 公网安备号（此字段已废弃，保留为兼容旧版本）
	Public *string `json:"public"`
}
