package file

import (
	"context"
	"io"
)

// Option 文件操作接口
type Option interface {
	// UploadFile 上传文件
	//   - file: 文件流
	//   - path: 文件路径
	//   - fileName: 文件名
	// Returns: 是否上传成功
	UploadFile(ctx context.Context, file io.Reader, path string, fileName string) (bool, error)

	// DeleteFiles 批量删除文件
	//   - fileNames: 文件名数组
	// Returns: 删除成功的文件名数组，删除失败的原因
	DeleteFiles(ctx context.Context, fileNames []string) ([]string, error)

	// MoveFile 移动文件
	//   - oldFileNames: 旧文件名数组
	//   - newGroupName: 要移动到的新的文件组（文件夹）名
	// Returns: 成功移动的文件的旧文件名（包括文件夹名），移动失败的原因
	MoveFile(ctx context.Context, oldFileNames []string, newGroupName string) ([]string, error)

	// IsExist 判断文件是否存在
	//   - fileName: 文件名
	// Returns: 是否存在
	IsExist(ctx context.Context, fileName string) bool
}
