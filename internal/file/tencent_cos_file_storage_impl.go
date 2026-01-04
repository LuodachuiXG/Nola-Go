package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nola-go/internal/file/config"
	"nola-go/internal/logger"
	"nola-go/internal/util"
	"path"
	"sync"

	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
)

var (
	_cosClient *cos.Client
	_config    *config.TencentCosConfig
	_mutex     sync.Mutex
)

// TencentCOSFileStorageImpl 腾讯云对象操作
type TencentCOSFileStorageImpl struct{}

// GetTencentCOSFileStorage 获取腾讯云对象存储操作实例
func GetTencentCOSFileStorage(config *config.TencentCosConfig) (*TencentCOSFileStorageImpl, error) {
	_mutex.Lock()
	defer _mutex.Unlock()

	// client 为 nil，并且 config 也为 nil
	if _cosClient == nil && config == nil {
		return nil, errors.New("腾讯云对象存储还未初始化，config 不能为 nil")
	}

	// config 不为 nil，更新 client
	if config != nil {
		_config = config
		protocol := "http"
		if config.Https {
			protocol = "https"
		}

		u, err := url.Parse(fmt.Sprintf("%s://%s.cos.%s.myqcloud.com", protocol, config.Bucket, config.Region))
		if err != nil {
			return nil, errors.New("腾讯云对象存储初始化失败，解析 URL 失败：" + err.Error())
		}

		b := &cos.BaseURL{BucketURL: u}
		_cosClient = cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  config.SecretId,
				SecretKey: config.SecretKey,
			},
		})
	}

	// config 为 nil，但是 client 不为 nil
	return &TencentCOSFileStorageImpl{}, nil
}

// UploadFile 上传文件
//   - file: 文件流
//   - path: 文件路径
//   - fileName: 文件名
//
// Returns: 是否上传成功
func (t TencentCOSFileStorageImpl) UploadFile(ctx context.Context, file io.Reader, filePath string, fileName string) (bool, error) {
	key := t.getFullKey(path.Join(filePath, fileName))

	_, err := _cosClient.Object.Put(ctx, key, file, nil)
	if err != nil {
		return false, fmt.Errorf("文件上传失败：%w", err)
	}
	return true, nil
}

// DeleteFiles 批量删除文件
//   - fileNames: 文件名数组（组名+文件名）
//
// Returns: 删除成功的文件名数组，删除失败的原因
func (t TencentCOSFileStorageImpl) DeleteFiles(ctx context.Context, fileNames []string) ([]string, error) {
	result := make([]string, 0)

	if len(fileNames) == 0 {
		return result, nil
	}

	// 构造批量删除的对象列表
	var obs []cos.Object
	for _, name := range fileNames {
		obs = append(obs, cos.Object{
			Key: t.getFullKey(name),
		})
	}

	opt := &cos.ObjectDeleteMultiOptions{
		Objects: obs,
	}

	ret, _, err := _cosClient.Object.DeleteMulti(ctx, opt)
	if err != nil {
		return result, fmt.Errorf("文件删除失败：%w", err)
	}

	// 记录成功删除的 Key
	for _, d := range ret.DeletedObjects {
		result = append(result, d.Key)
	}
	return result, nil
}

// MoveFile 移动文件
//   - oldFileNames: 旧文件名数组
//   - newGroupName: 要移动到的新的文件组（文件夹）名
//
// Returns: 成功移动的文件的旧文件名（包括文件夹名），移动失败的原因
func (t TencentCOSFileStorageImpl) MoveFile(ctx context.Context, oldFileNames []string, newGroupName string) ([]string, error) {
	successOldNames := make([]string, 0)
	for _, oldName := range oldFileNames {
		// 提取原文件名并拼接新路径
		baseName := path.Base(oldName)
		newName := path.Join(newGroupName, baseName)

		if ok, _ := t.copyFile(ctx, oldName, newName); ok {
			successOldNames = append(successOldNames, oldName)
		}
	}

	// 全部复制尝试完成后，批量删除原文件
	if len(successOldNames) > 0 {
		_, _ = t.DeleteFiles(ctx, successOldNames)
	}

	return successOldNames, nil
}

// IsExist 判断文件是否存在
//   - fileName: 文件名
//
// Returns: 是否存在
func (t TencentCOSFileStorageImpl) IsExist(ctx context.Context, fileName string) bool {
	key := t.getFullKey(fileName)
	// 使用 Head 判断对象是否存在
	ret, err := _cosClient.Object.IsExist(ctx, key)
	if err != nil {
		logger.Log.Error("腾讯云对象存储文件是否存在失败", zap.Error(err))
		return false
	}

	return ret
}

// TencentCOSUrl 拼接腾讯云对象存储文件链接地址
func TencentCOSUrl(config config.TencentCosConfig, fileName string, fileGroupPath *string) string {
	groupName := ""
	if fileGroupPath != nil {
		groupName = *fileGroupPath
	}

	pathStr := ""
	if config.Path != nil {
		pathStr = *config.Path
	}

	return "https://" + util.StringReplaceDoubleSlash(fmt.Sprintf(
		"%s.cos.%s.myqcloud.com/%s/%s/%s",
		config.Bucket,
		config.Region,
		pathStr,
		groupName,
		fileName,
	))
}

// getFullKey 获取完整的存储 Key
func (t TencentCOSFileStorageImpl) getFullKey(subPath string) string {
	basePath := ""
	if _config.Path != nil && *_config.Path != "" {
		basePath = *_config.Path
	}
	return util.StringFormatSlash(path.Join(basePath, subPath))
}

// copyFile 复制文件
func (t TencentCOSFileStorageImpl) copyFile(ctx context.Context, oldName string, newName string) (bool, error) {
	oldKey := t.getFullKey(oldName)
	newKey := t.getFullKey(newName)

	// 腾讯云复制源格式 <bucket>.cos.<region>.myqcloud.com/<key>
	sourceURL := fmt.Sprintf("%s.cos.%s.myqcloud.com/%s", _config.Bucket, _config.Region, oldKey)

	_, _, err := _cosClient.Object.Copy(ctx, newKey, sourceURL, nil)
	if err != nil {
		return false, fmt.Errorf("文件复制失败：%w", err)
	}
	return true, nil
}
