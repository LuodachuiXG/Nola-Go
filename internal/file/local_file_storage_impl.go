package file

import (
	"context"
	"errors"
	"io"
	"nola-go/internal/logger"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// UrlStoragePath URL 存储路径
const UrlStoragePath = "/upload"

// LocalStoragePath 本地存储路径
const LocalStoragePath = ".nola/" + UrlStoragePath

// LocalFileStorageImpl 本地存储实现
type LocalFileStorageImpl struct {
}

func NewLocalFileStorageImpl() *LocalFileStorageImpl {
	// 检查文件夹是否存在
	if _, err := os.Stat(LocalStoragePath); os.IsNotExist(err) {
		// 不存在，创建文件夹
		_ = os.MkdirAll(LocalStoragePath, 0755)
	}
	return &LocalFileStorageImpl{}
}

// UploadFile 上传文件
//   - file: 文件流
//   - path: 文件路径
//   - fileName: 文件名
//
// Returns: 是否上传成功
func (l LocalFileStorageImpl) UploadFile(_ context.Context, file io.Reader, path string, fileName string) (bool, error) {
	filePath := filepath.Join(LocalStoragePath, path)

	// 判断文件路径（文件夹是否存在）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件夹不存在，创建
		if err := os.MkdirAll(filePath, 0755); err != nil {
			return false, errors.New("无法创建文件夹：" + err.Error())
		}
	}

	// 创建文件
	fullPath := filepath.Join(filePath, fileName)
	newFile, err := os.Create(fullPath)
	if err != nil {
		return false, errors.New("无法创建文件：" + err.Error())
	}
	defer func() {
		_ = newFile.Close()
	}()

	// 拷贝数据流
	if _, err := io.Copy(newFile, file); err != nil {
		return false, errors.New("无法写入文件：" + err.Error())
	}
	return true, nil
}

// DeleteFiles 批量删除文件
//   - fileNames: 文件名数组
//
// Returns: 删除成功的文件名数组，删除失败的原因
func (l LocalFileStorageImpl) DeleteFiles(_ context.Context, fileNames []string) ([]string, error) {
	result := make([]string, 0)
	for _, fileName := range fileNames {
		fullPath := filepath.Join(LocalStoragePath, fileName)
		err := os.Remove(fullPath)
		if err == nil {
			result = append(result, fileName)
		} else {
			logger.Log.Error("无法删除文件：", zap.Error(err))
		}
	}
	return result, nil
}

// MoveFile 移动文件
//   - oldFileNames: 旧文件名数组
//   - newGroupName: 要移动到的新的文件组（文件夹）名
//
// Returns: 成功移动的文件的旧文件名（包括文件夹名），移动失败的原因
func (l LocalFileStorageImpl) MoveFile(_ context.Context, oldFileNames []string, newGroupName string) ([]string, error) {
	result := make([]string, 0)
	newGroupDirPath := filepath.Join(LocalStoragePath, newGroupName)

	// 确保新目录存在
	if _, err := os.Stat(newGroupDirPath); os.IsNotExist(err) {
		_ = os.MkdirAll(newGroupDirPath, 0755)
	}

	for _, oldFileName := range oldFileNames {
		oldFullPath := filepath.Join(LocalStoragePath, oldFileName)

		// 检查原文件是否存在
		if _, err := os.Stat(oldFullPath); err == nil {
			// 获取文件名
			baseName := filepath.Base(oldFileName)
			newFullPath := filepath.Join(newGroupDirPath, baseName)

			// 移动文件
			err := os.Rename(oldFullPath, newFullPath)
			if err == nil {
				result = append(result, oldFileName)
				// 检查父目录是否为空，为空则删除
				dir := filepath.Dir(oldFullPath)
				if isDirEmpty(dir) && dir != LocalStoragePath {
					_ = os.Remove(dir)
				}
			} else {
				logger.Log.Error("无法移动文件：", zap.Error(err))
			}
		}
	}

	return result, nil
}

// IsExist 判断文件是否存在
//   - fileName: 文件名
//
// Returns: 是否存在
func (l LocalFileStorageImpl) IsExist(_ context.Context, fileName string) bool {
	fullPath := filepath.Join(LocalStoragePath, fileName)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// isDirEmpty 判断目录是否为空
//   - name: 目录地址
func isDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// 读取目录中的第一个条目
	// Readdirnames(n) 如果 n > 0，则最多返回 n 个条目
	_, err = f.Readdirnames(1)

	// 如果返回的错误是 io.EOF，说明目录下没有任何文件/子目录
	if errors.Is(err, io.EOF) {
		return true
	}

	// 如果没有错误，说明至少找到了一个条目，文件夹不为空
	// 如果有其他错误，则返回错误
	return false
}
