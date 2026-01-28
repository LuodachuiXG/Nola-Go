package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"nola-go/internal/logger"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// IsDirExist 文件夹是否存在
func IsDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err)
	}

	// 确认是目录
	if !info.IsDir() {
		// 不是目录
		return false
	}

	return true
}

// IsDir 判断是否为文件夹
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		logger.Log.Error("判断是否为文件夹失败", zap.Error(err))
		return false
	}
	return info.IsDir()
}

// CreateFolderZip 将一个文件夹创建 Zip 压缩文件
// 只处理第一层文件，不处理子文件夹
//   - folderPath: 要压缩的文件夹路径
//   - zipFileName: 压缩包名称
func CreateFolderZip(folderPath, zipFileName string) error {
	// 检查文件夹是否存在
	info, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件夹不存在
			return errors.New(fmt.Sprintf("文件夹 [%s] 不存在", folderPath))
		}
		return err
	}

	// 检查是否为文件夹
	if !info.IsDir() {
		return errors.New(fmt.Sprintf("[%s] 不是文件夹", folderPath))
	}

	// 创建 Zip
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}

	defer func() {
		_ = zipFile.Close()
	}()

	// 创建 Zip 输出流
	zipWriter := zip.NewWriter(zipFile)

	// 遍历文件夹并添加到 Zip
	err = filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录，只处理文件
		if info.IsDir() {
			return nil
		}

		// 跳过隐藏文件
		if info.Name()[0] == '.' {
			return nil
		}

		// 计算相对于根目录的路径
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// 将路径分隔符统一替换为正斜杠
		relPath = filepath.ToSlash(relPath)

		// 创建 Zip 实体
		zipEntry, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		zipEntry.Name = relPath
		// 压缩文件
		zipEntry.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(zipEntry)
		if err != nil {
			return err
		}

		// 读取源文件并写入 Zip
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		_ = zipFile.Close()
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	return zipFile.Close()
}
