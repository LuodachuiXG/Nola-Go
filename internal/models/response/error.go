package response

import (
	"errors"
	"fmt"
	"nola-go/internal/models/enum"
)

var (
	ServerError = errors.New("未知错误，请检查服务器日期")
)

// FileStorageNotConfiguredError 文件组策略还未配置异常
func FileStorageNotConfiguredError(mode enum.FileStorageMode) error {
	return errors.New(fmt.Sprintf("文件存储策略 [%s] 还未配置", mode))
}
