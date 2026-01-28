package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"nola-go/internal/file"
	"nola-go/internal/file/config"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"
	"path/filepath"
	"slices"
	"time"

	"go.uber.org/zap"
)

type FileService struct {
	fileRepo repository.FileRepository

	// 本地存储
	localStorage file.Option
	// 腾讯云对象存储
	tencentCOS file.Option
	// 腾讯云对象存储配置
	tencentConfig *config.TencentCosConfig
}

func NewFileService(fileRepo repository.FileRepository) *FileService {
	return &FileService{
		fileRepo:      fileRepo,
		localStorage:  file.LocalFileStorageImpl{},
		tencentCOS:    nil,
		tencentConfig: nil,
	}
}

// InitFileStorageMode 初始化文件存储方式（除了本地存储）
//   - mode: 文件存储方式（除了本地存储 LOCAL）
//
// Returns: 已经初始化或初始化成功的实例，初始化失败返回 nil。
func (s *FileService) InitFileStorageMode(ctx context.Context, mode enum.FileStorageMode) (file.Option, error) {
	switch mode {
	// 腾讯云对象存储
	case enum.FileStorageModeTencentCOS:
		if s.tencentCOS != nil {
			// 腾讯云对象已经初始化，直接返回对象
			return s.tencentCOS, nil
		}

		// 腾讯云对象未初始化，先尝试获取腾讯云对象配置信息
		tConfig, err := s.fileRepo.GetFileStorageConfig(ctx, mode)
		if err != nil || util.StringIsNilOrBlank(tConfig) {
			logger.Log.Error("获取腾讯云对象存储配置失败", zap.Error(err))
			return nil, response.ServerError
		}

		if s.tencentCOS == nil {
			s.tencentConfig = &config.TencentCosConfig{}
		}
		// 反序列化腾讯云对象存储 JSON
		err = util.FromJsonString(tConfig, s.tencentConfig)
		if err != nil {
			logger.Log.Error("解析腾讯云对象存储配置失败", zap.Error(err))
			return nil, response.ServerError
		}

		// 设置腾讯云对象存储
		client, err := file.GetTencentCOSFileStorage(s.tencentConfig)

		if err != nil {
			logger.Log.Error("获取腾讯云对象存储实例失败", zap.Error(err))
			return nil, response.ServerError
		}

		s.tencentCOS = client

		return s.tencentCOS, nil
	}

	return nil, nil
}

// SetTencentCOS 设置腾讯云对象存储配置
//   - config: 腾讯云对象存储配置
func (s *FileService) SetTencentCOS(ctx context.Context, config *config.TencentCosConfig) (bool, error) {
	ret, err := s.fileRepo.SetFileStorageConfig(ctx, enum.FileStorageModeTencentCOS, *util.ToJsonString(config))

	if err != nil {
		logger.Log.Error("设置腾讯云对象存储失败", zap.Error(err))
		return false, response.ServerError
	}

	if ret {
		// 设置腾讯云对象存储成功，重新设置腾讯云对象存储变量
		s.tencentCOS = nil
		_, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)
		if err != nil {
			return false, err
		}
	}

	return ret, nil
}

// DeleteTencentCOS 删除腾讯云对象存储配置
func (s *FileService) DeleteTencentCOS(ctx context.Context) (bool, error) {
	// 先判断腾讯云对线存储下是否还有文件
	count, err := s.fileRepo.GetFileCountByMode(ctx, enum.FileStorageModeTencentCOS)
	if err != nil {
		logger.Log.Error("获取腾讯云对象存储文件数量失败", zap.Error(err))
		return false, response.ServerError
	}

	if count > 0 {
		return false, errors.New(fmt.Sprintf("腾讯云对象存储策略下还有 %d 个文件，无法删除", count))
	}

	ret, err := s.fileRepo.DeleteFileStorageConfig(ctx, enum.FileStorageModeTencentCOS)
	if err != nil {
		logger.Log.Error("删除腾讯云对象存储失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// GetTencentCOS 获取腾讯云对象存储配置
func (s *FileService) GetTencentCOS(ctx context.Context) (*config.TencentCosConfig, error) {
	configStr, err := s.fileRepo.GetFileStorageConfig(ctx, enum.FileStorageModeTencentCOS)
	if err != nil {
		logger.Log.Error("获取腾讯云对象存储配置失败", zap.Error(err))
		return nil, response.ServerError
	}

	// 反序列化
	var c *config.TencentCosConfig
	err = util.FromJsonString(configStr, &c)
	if err != nil {
		logger.Log.Error("反序列化腾讯云对象存储配置失败", zap.Error(err))
		return nil, response.ServerError
	}

	return c, nil
}

// GetModes 获取所有已经设置过的存储策略
// 默认包含本地存储（LOCAL）
func (s *FileService) GetModes(ctx context.Context) ([]enum.FileStorageMode, error) {
	ret := []enum.FileStorageMode{enum.FileStorageModeLocal}

	modes, err := s.fileRepo.GetModes(ctx)

	if err != nil {
		logger.Log.Error("获取文件存储策略失败", zap.Error(err))
		return nil, response.ServerError
	}

	for _, mode := range modes {
		ret = append(ret, mode)
	}

	return ret, nil
}

// AddFileGroup 添加文件组
func (s *FileService) AddFileGroup(ctx context.Context, group request.FileGroupAddRequest) (*models.FileGroup, error) {
	// 如果文件存储方式不是本地存储，就先检查对应的存储方式是否应设置
	if group.StorageMode != enum.FileStorageModeLocal {
		isSet, err := s.IsModeSet(ctx, group.StorageMode)
		if err != nil {
			return nil, err
		}

		if !isSet {
			return nil, errors.New(fmt.Sprintf("文件存储策略 [%s] 还未设置", group.StorageMode))
		}
	}

	// 添加前先判断相同存储方式下，文件组名和文件组路径是否重复
	g, err := s.GetFileGroupByDisplayName(ctx, group.StorageMode, group.DisplayName)
	if err != nil {
		return nil, err
	}
	if g != nil {
		return nil, errors.New(fmt.Sprintf("文件组名 [%s] 已经存在", group.DisplayName))
	}

	// 再判断路径是否重复
	p, err := s.GetFileGroupByPath(ctx, group.StorageMode, group.Path)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return nil, errors.New(fmt.Sprintf("文件组路径 [%s] 已经存在", group.Path))
	}

	// 添加文件组
	ret, err := s.fileRepo.AddFileGroup(ctx, group)

	if err != nil {
		logger.Log.Error("添加文件组失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// DeleteFileGroup 删除文件组
func (s *FileService) DeleteFileGroup(ctx context.Context, fileGroupId uint) (bool, error) {
	// 删除前先判断当前文件组下是否还有文件
	count, err := s.fileRepo.GetFileCountByGroup(ctx, fileGroupId)
	if err != nil {
		logger.Log.Error("获取文件组文件数量失败", zap.Error(err))
		return false, response.ServerError
	}

	if count > 0 {
		return false, errors.New(fmt.Sprintf("文件组下还有 %d 个文件，无法删除", count))
	}

	ret, err := s.fileRepo.DeleteFileGroup(ctx, fileGroupId)
	if err != nil {
		logger.Log.Error("删除文件组失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdateFileGroup 修改文件组
// 文件组存储方式不能更改
func (s *FileService) UpdateFileGroup(ctx context.Context, fileGroup *request.FileGroupUpdateRequest) (bool, error) {
	// 先获取要修改的文件组
	oldGroup, err := s.GetFileGroupById(ctx, fileGroup.FileGroupId)
	if err != nil {
		return false, err
	}

	if oldGroup == nil {
		return false, errors.New(fmt.Sprintf("文件组 [%d] 不存在", fileGroup.FileGroupId))
	}

	// 如果文件组名已经存在，并且不是自己
	fg, err := s.GetFileGroupByDisplayName(ctx, oldGroup.StorageMode, fileGroup.DisplayName)
	if err != nil {
		return false, err
	}

	if fg != nil && fg.FileGroupId != fileGroup.FileGroupId {
		return false, errors.New(fmt.Sprintf("文件组名 [%s] 已经存在", fileGroup.DisplayName))
	}

	// 修改文件组
	ret, err := s.fileRepo.UpdateFileGroup(ctx, *fileGroup)
	if err != nil {
		logger.Log.Error("修改文件组失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// GetFileGroupById 根据文件组 ID 获取文件组
func (s *FileService) GetFileGroupById(ctx context.Context, fileGroupId uint) (*models.FileGroup, error) {
	ret, err := s.fileRepo.GetFileGroupById(ctx, fileGroupId)
	if err != nil {
		logger.Log.Error("获取文件组失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// GetFileGroupsByMode 根据文件存储方式获取文件组
//   - mode: 文件存储方式（nil 获取所有文件组）
func (s *FileService) GetFileGroupsByMode(ctx context.Context, mode *enum.FileStorageMode) ([]*models.FileGroup, error) {
	ret, err := s.fileRepo.GetFileGroupsByMode(ctx, mode)
	if err != nil {
		logger.Log.Error("获取文件组失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// GetFileGroupByDisplayName 根据文件存储方式和文件组名获取文件组
func (s *FileService) GetFileGroupByDisplayName(ctx context.Context, mode enum.FileStorageMode, displayName string) (*models.FileGroup, error) {
	ret, err := s.GetFileGroupsByMode(ctx, &mode)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(ret, func(group *models.FileGroup) bool {
		return group.DisplayName == displayName
	})

	if index != -1 {
		return ret[index], nil
	}
	return nil, nil
}

// GetFileGroupByPath 根据文件存储方式和文件组路径获取文件组
func (s *FileService) GetFileGroupByPath(ctx context.Context, mode enum.FileStorageMode, path string) (*models.FileGroup, error) {
	ret, err := s.GetFileGroupsByMode(ctx, &mode)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(ret, func(group *models.FileGroup) bool {
		return group.Path == path
	})

	if index != -1 {
		return ret[index], nil
	}
	return nil, nil
}

// UploadFile 上传文件
//   - ctx: 上下文
//   - file: 文件流
//   - fileName: 文件名
//   - mode: 文件存储方式
//   - groupId: 文件组 ID
//   - length: 文件长度
func (s *FileService) UploadFile(
	ctx context.Context,
	fileIO io.Reader,
	fileName string,
	mode enum.FileStorageMode,
	groupId *uint,
	length *int64,
) (*response.FileResponse, error) {
	var fileGroup *models.FileGroup = nil

	if groupId != nil {
		// 先尝试获取文件组
		fg, err := s.GetFileGroupById(ctx, *groupId)
		if err != nil {
			return nil, err
		}

		if fg == nil || fg.StorageMode != mode {
			return nil, errors.New(fmt.Sprintf("文件组 [%d] 不存在", *groupId))
		}

		fileGroup = fg
	}

	// 如果存储方式不为本地存储，就先查看对应的存储方式是否已经配置
	if mode != enum.FileStorageModeLocal {
		// 先尝试初始化存储方式
		ret, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)

		if err != nil {
			return nil, err
		}

		if ret == nil {
			// 腾讯云存储方式还未配置
			return nil, response.FileStorageNotConfiguredError(enum.FileStorageModeTencentCOS)
		}
	}

	// 查看当前文件名是否已经存在
	f, err := s.fileRepo.GetFile(ctx, fileName, groupId, mode)
	if err != nil {
		logger.Log.Error("获取文件失败", zap.Error(err))
		return nil, response.ServerError
	}

	// 最终文件名
	actualFileName := fileName

	if f != nil {
		// 文件已经存在，给文件名加上 5 个随机数字或字母
		actualFileName = util.StringFileNameAddRandomSuffix(actualFileName)
	}

	// 文件路径
	path := ""
	if fileGroup != nil {
		// 文件组不为空，用文件组的路径
		path = fileGroup.Path
	}

	// 上传结果
	ret := false
	switch mode {
	case enum.FileStorageModeLocal:
		// 本地存储
		ret, _ = s.localStorage.UploadFile(ctx, fileIO, path, actualFileName)
	case enum.FileStorageModeTencentCOS:
		// 腾讯云存储
		// 上传文件
		// 按官方文档说腾讯云上传的文件最好给定文件的大小，不然会腾讯云对象存储会计算流长度
		// 这会导致耗时操作，并且会占用不必要的内存
		// 但是，请求头 Content-Length 传过来 fileLength 长度似乎和腾讯云对象存储的算出的不一致，
		// 这会导致腾讯云对象存储报错，导致上传失败，这是待解决问题！
		ret, _ = s.tencentCOS.UploadFile(ctx, fileIO, path, actualFileName)
	}

	// 如果文件上传成功，将新文件插入数据库
	if ret {
		f, err := s.fileRepo.AddFile(ctx, models.File{
			FileGroupId: groupId,
			DisplayName: actualFileName,
			Size:        *util.DefaultPtr(length, 1),
			StorageMode: mode,
			CreateTime:  time.Now().UnixMilli(),
		})

		if err != nil {
			logger.Log.Error("添加文件失败", zap.Error(err))
			return nil, response.ServerError
		}

		// 文件响应
		fileRes := &response.FileResponse{
			FileGroupId: groupId,
			DisplayName: actualFileName,
			Size:        *util.DefaultPtr(length, 1),
			StorageMode: mode,
		}

		if f != nil {
			fileRes.FileId = f.FileId
			fileRes.CreateTime = f.CreateTime
		} else {
			//fileRes.FileId = -1
			fileRes.CreateTime = time.Now().UnixMilli()
		}

		if fileGroup != nil {
			fileRes.FileGroupName = &fileGroup.DisplayName
		}

		switch mode {
		case enum.FileStorageModeLocal:
			fileRes.Url = util.StringFormatSlash(fmt.Sprintf("/upload/%s/%s", path, actualFileName))
		case enum.FileStorageModeTencentCOS:
			var groupPath *string = nil
			if fileGroup != nil {
				groupPath = &fileGroup.Path
			}
			fileRes.Url = file.TencentCOSUrl(*s.tencentConfig, actualFileName, groupPath)
		}

		return fileRes, nil
	}

	return nil, nil
}

// UploadFileRecord 添加上传文件记录
func (s *FileService) UploadFileRecord(ctx context.Context, record request.FileRecordRequest) (*response.FileResponse, error) {
	currentStorageMode := enum.FileStorageModeLocal
	if record.StorageMode != nil {
		currentStorageMode = *record.StorageMode
	}

	var fileGroup *models.FileGroup = nil

	if record.FileGroupId != nil {
		// 文件组 ID 不为空
		// 先尝试获取文件组
		fg, err := s.GetFileGroupById(ctx, *record.FileGroupId)
		if err != nil {
			return nil, err
		}

		if fg == nil || fg.StorageMode != currentStorageMode {
			return nil, errors.New(fmt.Sprintf("文件组 [%d] 不存在", record.FileGroupId))
		}

		fileGroup = fg
	}

	// 如果存储方式不为本地存储，就先查看对应的存储方式是否已经配置
	if currentStorageMode != enum.FileStorageModeLocal {
		// 先尝试初始化存储方式
		ret, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)
		if err != nil {
			return nil, err
		}
		if ret == nil {
			// 腾讯云还未配置
			return nil, response.FileStorageNotConfiguredError(enum.FileStorageModeTencentCOS)
		}
	}

	// 查看当前文件名是否已经存在
	f, err := s.fileRepo.GetFile(ctx, record.Name, record.FileGroupId, currentStorageMode)
	if err != nil {
		logger.Log.Error("获取文件失败", zap.Error(err))
		return nil, response.ServerError
	}
	if f != nil {
		// 文件已经存在
		// 这里和上面的上传接口此处处理逻辑不一样，是因为此接口是针对已经上传过的文件，添加文件记录
		// 所以这里如果文件名重复，直接抛出异常即可
		return nil, errors.New(fmt.Sprintf("文件名 [%s] 已经存在", record.Name))
	}

	// 将文件信息添加到数据库
	newFile := models.File{
		FileGroupId: record.FileGroupId,
		DisplayName: record.Name,
		Size:        record.Size,
		StorageMode: currentStorageMode,
		CreateTime:  time.Now().UnixMilli(),
	}

	newFileResult, err := s.fileRepo.AddFile(ctx, newFile)
	if err != nil {
		logger.Log.Error("添加文件失败", zap.Error(err))
		return nil, response.ServerError
	}

	// 返回结果
	ret := &response.FileResponse{
		FileGroupId: record.FileGroupId,
		DisplayName: record.Name,
		Size:        record.Size,
		StorageMode: currentStorageMode,
	}

	if newFileResult != nil {
		ret.FileId = newFileResult.FileId
		ret.CreateTime = newFileResult.CreateTime
	} else {
		//ret.FileId = -1
		ret.CreateTime = time.Now().UnixMilli()
	}

	if fileGroup != nil {
		ret.FileGroupName = &fileGroup.DisplayName
	}

	path := ""
	if fileGroup != nil {
		path = fileGroup.Path
	}

	switch currentStorageMode {
	case enum.FileStorageModeLocal:
		ret.Url = util.StringFormatSlash(fmt.Sprintf("/upload/%s/%s", path, record.Name))
	case enum.FileStorageModeTencentCOS:
		var groupPath *string = nil
		if fileGroup != nil {
			groupPath = &fileGroup.Path
		}
		ret.Url = file.TencentCOSUrl(*s.tencentConfig, record.Name, groupPath)
	}

	return ret, nil
}

// DeleteFiles 根据文件 ID 数组删除文件
//   - ids: 文件 ID 数组
//
// Returns: 删除成功的文件 ID 数组
func (s *FileService) DeleteFiles(ctx context.Context, ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return []uint{}, nil
	}

	// 根据文件 ID 获取所有文件
	files, err := s.fileRepo.GetFileWithGroupByIds(ctx, ids)
	if err != nil {
		logger.Log.Error("获取文件失败", zap.Error(err))
		return nil, response.ServerError
	}

	// 将要删除的文件封装成文件索引数据 [FileIndex]
	fileIndexes := util.Map(files, func(f *models.FileWithGroup) *models.FileIndex {
		path := ""
		if f.FileGroupPath != nil {
			path = *f.FileGroupPath
		}

		return &models.FileIndex{
			FileId:      &f.FileId,
			Name:        fmt.Sprintf("%s/%s", path, f.FileName),
			StorageMode: f.StorageMode,
		}
	})

	// 删除文件，获取成功删除的文件索引数据类
	deleteResult, err := s.DeleteFilesByFileIndexes(ctx, fileIndexes)
	if err != nil {
		return nil, err
	}

	if len(deleteResult) == len(ids) {
		// 成功删除的文件数量和要删除的文件数量相同，返回要删除的文件 ID 数组
		return ids, nil
	}

	// 返回删除成功的文件 ID
	return util.Map(deleteResult, func(f *models.FileIndex) uint {
		return *f.FileId
	}), nil
}

// DeleteFilesByFileIndexes 根据文件索引删除文件
//   - fileIndexes: 文件索引数组
//
// Returns: 删除成功的文件索引数组
func (s *FileService) DeleteFilesByFileIndexes(ctx context.Context, fileIndexes []*models.FileIndex) ([]*models.FileIndex, error) {
	if len(fileIndexes) == 0 {
		return []*models.FileIndex{}, nil
	}

	// 将不同存储形式的文件分离
	var localFileIndexes []*models.FileIndex
	var tencentFileIndexes []*models.FileIndex
	for _, fileIndex := range fileIndexes {
		switch fileIndex.StorageMode {
		case enum.FileStorageModeLocal:
			localFileIndexes = append(localFileIndexes, fileIndex)
		case enum.FileStorageModeTencentCOS:
			tencentFileIndexes = append(tencentFileIndexes, fileIndex)
		}
	}

	// 成功删除的文件索引数组
	var resultFileIndexes []*models.FileIndex
	if len(tencentFileIndexes) > 0 {
		// 先尝试初始化存储方式
		r, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)
		if err != nil {
			return nil, err
		}
		if r == nil {
			return nil, response.FileStorageNotConfiguredError(enum.FileStorageModeTencentCOS)
		}

		// 删除腾讯云对象存储文件
		tencentDeleteResult, err := s.tencentCOS.DeleteFiles(ctx,
			util.Map(tencentFileIndexes, func(index *models.FileIndex) string {
				return index.Name
			}),
		)

		if err != nil {
			logger.Log.Error("删除腾讯云对象存储文件失败", zap.Error(err))
			return nil, response.ServerError
		}

		if len(tencentDeleteResult) == len(tencentFileIndexes) {
			// 如果成功删除的文件数目与请求删除的文件数目相等，就将所有请求删除的文件都加入删除成功结果数组
			resultFileIndexes = append(resultFileIndexes, tencentFileIndexes...)
		} else {
			// 成功删除的文件数目与请求删除的文件数目不想等，只把成功删除的文件加入删除成功结果数组
			for _, result := range tencentDeleteResult {

				firstIndex := slices.IndexFunc(tencentFileIndexes, func(index *models.FileIndex) bool {
					return index.Name == result
				})

				if firstIndex == -1 {
					continue
				}

				resultFileIndexes = append(resultFileIndexes, tencentFileIndexes[firstIndex])
			}
		}
	}

	// 如果本地存储文件不为空，就删除本地文件
	if len(localFileIndexes) > 0 {
		localDeleteResult, err := s.localStorage.DeleteFiles(ctx,
			util.Map(localFileIndexes, func(index *models.FileIndex) string {
				return index.Name
			}),
		)

		if err != nil {
			logger.Log.Error("删除本地文件失败", zap.Error(err))
			return nil, response.ServerError
		}

		if len(localDeleteResult) == len(localFileIndexes) {
			// 如果成功删除的文件数目与请求删除的文件数目相等，就将所有请求删除的文件都加入删除成功结果数组
			resultFileIndexes = append(resultFileIndexes, localFileIndexes...)
		} else {
			// 成功删除的文件数目与请求删除的文件数目不相等，只把成功删除的文件加入删除成功结果数组
			for _, result := range localDeleteResult {
				firstIndex := slices.IndexFunc(localFileIndexes, func(index *models.FileIndex) bool {
					return index.Name == result
				})

				if firstIndex == -1 {
					continue
				}

				resultFileIndexes = append(resultFileIndexes, localFileIndexes[firstIndex])
			}
		}
	}

	if len(resultFileIndexes) > 0 {
		// 如果删除成功的文件不为空，就删除数据库中的文件记录
		_, err := s.deleteDatabaseFilesByFileIndexes(ctx, resultFileIndexes)
		if err != nil {
			return nil, err
		}
	}

	// 返回删除成功的结果集
	if resultFileIndexes == nil {
		return []*models.FileIndex{}, nil
	}
	return resultFileIndexes, nil
}

// MoveFiles 移动文件
//   - fileMoveRequest: 文件移动请求
//
// Returns: 移动成功的旧文件的地址数组
func (s *FileService) MoveFiles(ctx context.Context, fileMoveRequest *request.FileMoveRequest) ([]string, error) {
	if fileMoveRequest == nil || len(fileMoveRequest.FileIds) == 0 {
		return []string{}, nil
	}

	// 先获取所有要移动的文件
	files, err := s.fileRepo.GetFileWithGroupByIds(ctx, fileMoveRequest.FileIds)
	if err != nil {
		logger.Log.Error("获取文件失败", zap.Error(err))
		return []string{}, response.ServerError
	}
	if len(files) == 0 {
		return []string{}, nil
	}

	var newFileGroup *models.FileGroup = nil

	// 判断是否所有的文件都属于同一存储方式
	firstFile := files[0]
	isSameStorageMode := true
	for _, f := range files {
		if f.StorageMode != firstFile.StorageMode {
			isSameStorageMode = false
			break
		}
	}

	if len(files) > 1 && !isSameStorageMode {
		// 待移动的文件不是同属于一个存储方式
		return []string{}, errors.New("待移动的文件必须属于同一存储策略")
	}

	// 判断要移动进的新的文件组是否存在
	if fileMoveRequest.NewFileGroupId != nil && *fileMoveRequest.NewFileGroupId > 0 {
		group, err := s.GetFileGroupById(ctx, *fileMoveRequest.NewFileGroupId)
		if err != nil {
			return []string{}, err
		}
		newFileGroup = group

		if newFileGroup == nil || newFileGroup.StorageMode != firstFile.StorageMode {
			// 文件组为空，或者存储方式与待移动的文件不一致
			// 话句话说，如果当前文件已经存在目标文件组中，就不进行操作。
			return []string{}, errors.New(fmt.Sprintf("文件组 [%d] 不存在", *fileMoveRequest.NewFileGroupId))
		}
	}

	// 待移动的文件名（包括文件组路径）
	var waitForMoveFileNames []string
	for _, fg := range files {
		if newFileGroup == nil || fg.FileGroupId == nil || (*fg.FileGroupId != newFileGroup.FileGroupId) {
			// 要移动的文件，不在目标文件组中，就加入待移动的文件数组中

			path := ""
			if fg.FileGroupPath != nil {
				path = *fg.FileGroupPath
			}

			waitForMoveFileNames = append(
				waitForMoveFileNames,
				util.StringFormatSlash(fmt.Sprintf("%s/%s", path, fg.FileName)),
			)
		}
	}

	// 成功移动的文件名
	var movedFileNames []string
	path := ""
	if newFileGroup != nil {
		path = newFileGroup.Path
	}
	switch firstFile.StorageMode {
	case enum.FileStorageModeLocal:
		// 本地文件移动
		ret, err := s.localStorage.MoveFile(ctx, waitForMoveFileNames, path)
		if err != nil {
			logger.Log.Error("移动本地文件失败", zap.Error(err))
			return []string{}, response.ServerError
		}
		// 将成功移动的文件名加入结果数组
		movedFileNames = append(movedFileNames, ret...)
	case enum.FileStorageModeTencentCOS:
		// 腾讯云对象存储文件移动
		// 先尝试初始化腾讯云对象存储方式
		// 先尝试初始化存储方式
		tc, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)

		if err != nil {
			return []string{}, err
		}

		if tc == nil {
			return []string{}, response.FileStorageNotConfiguredError(enum.FileStorageModeTencentCOS)
		}

		ret, err := s.tencentCOS.MoveFile(ctx, waitForMoveFileNames, path)
		if err != nil {
			logger.Log.Error("移动腾讯云对象存储文件失败", zap.Error(err))
			return []string{}, response.ServerError
		}
		// 将成功移动的文件名加入结果数组
		movedFileNames = append(movedFileNames, ret...)
	}

	// 修改成功移动的文件的文件组
	var newFiles []models.File
	for _, name := range movedFileNames {
		// 寻找成功移动的文件
		f := *util.Find(files, func(ff *models.FileWithGroup) bool {
			path := ""
			if ff.FileGroupPath != nil {
				path = *ff.FileGroupPath
			}

			return fmt.Sprintf("%s/%s", path, ff.FileName) == name
		})

		if f != nil {
			var groupId *uint = nil
			if newFileGroup != nil {
				groupId = &newFileGroup.FileGroupId
			}
			newFiles = append(
				newFiles,
				models.File{
					FileId: f.FileId,
					// 关键修改点
					FileGroupId: groupId,
					DisplayName: f.FileName,
					Size:        f.Size,
					StorageMode: f.StorageMode,
					CreateTime:  f.CreateTime,
				},
			)
		}
	}

	// 修改成功移动的文件的文件组
	_, err = s.fileRepo.UpdateFiles(ctx, newFiles)
	if err != nil {
		logger.Log.Error("修改文件组失败", zap.Error(err))
		return []string{}, response.ServerError
	}

	return util.DefaultEmptySlice(movedFileNames), nil
}

// GetFiles 获取文件
//   - page: 当前页
//   - size: 每页条数
//   - sort: 排序方式
//   - mode: 文件存储方式
//   - groupId: 文件组 ID
//   - key: 关键字
func (s *FileService) GetFiles(
	ctx context.Context,
	page, size int,
	sort *enum.FileSort,
	mode *enum.FileStorageMode,
	groupId *uint,
	key *string,
) (*models.Pager[response.FileResponse], error) {
	var fileResponse []*response.FileResponse
	// 先分页获取文件和文件组分页数据
	fileWithGroupPager, err := s.fileRepo.GetFileWithGroups(ctx, page, size, sort, mode, groupId, key)
	if err != nil {
		logger.Log.Error("获取文件和文件组失败", zap.Error(err))
		return nil, response.ServerError
	}

	for _, fg := range fileWithGroupPager.Data {
		res := &response.FileResponse{
			FileId:        fg.FileId,
			FileGroupId:   fg.FileGroupId,
			FileGroupName: fg.FileGroupName,
			DisplayName:   fg.FileName,
			Size:          fg.Size,
			StorageMode:   fg.StorageMode,
			CreateTime:    fg.CreateTime,
		}

		groupPath := ""
		if fg.FileGroupPath != nil {
			groupPath = *fg.FileGroupPath
		}

		switch fg.StorageMode {
		case enum.FileStorageModeLocal:
			// 本次存储，返回相对地址
			res.Url = util.StringFormatSlash(fmt.Sprintf("%s/%s/%s", file.UrlStoragePath, groupPath, fg.FileName))
		case enum.FileStorageModeTencentCOS:
			// 腾讯云对象存储，返回绝对地址
			// 先尝试初始化腾讯云对象存储
			option, err := s.InitFileStorageMode(ctx, enum.FileStorageModeTencentCOS)
			if err != nil {
				return nil, err
			}
			if option == nil {
				return nil, response.FileStorageNotConfiguredError(enum.FileStorageModeTencentCOS)
			}

			res.Url = file.TencentCOSUrl(*s.tencentConfig, fg.FileName, fg.FileGroupPath)
		}

		fileResponse = append(fileResponse, res)
	}

	return &models.Pager[response.FileResponse]{
		Page:       fileWithGroupPager.Page,
		Size:       fileWithGroupPager.Size,
		Data:       fileResponse,
		TotalData:  fileWithGroupPager.TotalData,
		TotalPages: fileWithGroupPager.TotalPages,
	}, nil
}

// FileCount 文件数量
func (s *FileService) FileCount(ctx context.Context) (int64, error) {
	count, err := s.fileRepo.GetFileCount(ctx)
	if err != nil {
		logger.Log.Error("获取文件数量失败", zap.Error(err))
		return 0, response.ServerError
	}
	return count, nil
}

// IsModeSet 检测文件存储方式是否已设置
func (s *FileService) IsModeSet(ctx context.Context, mode enum.FileStorageMode) (bool, error) {
	m, err := s.fileRepo.GetFileStorageConfig(ctx, mode)
	if err != nil {
		logger.Log.Error("获取文件存储方式配置失败", zap.Error(err))
		return false, response.ServerError
	}
	return m != nil, nil
}

// deleteDatabaseFilesByFileIndexes 根据文件索引删除数据库中总的文件记录
//   - fileIndexes: 文件索引数组
func (s *FileService) deleteDatabaseFilesByFileIndexes(ctx context.Context, fileIndexes []*models.FileIndex) (bool, error) {

	if len(fileIndexes) == 0 {
		return false, nil
	}

	// 映射出本次要删除的文件的文件组地址（如果有）和文件索引[地址:文件索引]
	var pathIndexes []*models.Pair[string, *models.FileIndex]
	for _, fileIndex := range fileIndexes {
		dir := filepath.Dir(fileIndex.Name)

		if len(dir) > 0 && dir[0] != filepath.Separator {
			// 如果文件组地址没有斜杠开头，则加上斜杠，防止匹配不到数据库文件组数据
			dir = string(filepath.Separator) + dir
		}
		pathIndexes = append(pathIndexes, &models.Pair[string, *models.FileIndex]{
			First:  dir,
			Second: fileIndex,
		})
	}

	// 过滤出有文件组的文件
	hasPath := util.Filter(pathIndexes, func(pair *models.Pair[string, *models.FileIndex]) bool {
		// 过滤掉没有文件组的文件
		return s.pathHasFileGroup(pair.First)
	})

	// 获取本次要删除的文件涉及到的所有文件组
	fileGroups, err := s.fileRepo.GetFileGroupsByPath(ctx,
		util.Map(hasPath,
			func(pair *models.Pair[string, *models.FileIndex]) string {
				return pair.First
			},
		),
	)

	// 文件组地址和 ID 映射
	fileGroupMap := make(map[string]*uint)
	for _, fg := range fileGroups {
		fileGroupMap[fg.Path] = &fg.FileGroupId
	}

	if err != nil {
		logger.Log.Error("获取文件组失败", zap.Error(err))
		return false, response.ServerError
	}

	// 要删除的文件 [文件组 ID (如果有) : 文件名]
	var localFiles []*models.Pair[*uint, string]
	var tencentFiles []*models.Pair[*uint, string]

	for _, index := range pathIndexes {
		pair := &models.Pair[*uint, string]{
			First: nil,
			// 获取文件名
			Second: filepath.Base(index.Second.Name),
		}
		if s.pathHasFileGroup(index.First) {
			// 当前文件有文件组
			pair.First = fileGroupMap[index.First]
		}

		// 根据不同的存储策略，加入对应的数组
		switch index.Second.StorageMode {
		case enum.FileStorageModeLocal:
			localFiles = append(localFiles, pair)
		case enum.FileStorageModeTencentCOS:
			tencentFiles = append(tencentFiles, pair)
		}
	}

	if len(localFiles) > 0 {
		_, err := s.fileRepo.DeleteFileByGroupIdAndName(ctx, localFiles, enum.FileStorageModeLocal)
		if err != nil {
			logger.Log.Error("删除数据库中的本地文件记录失败", zap.Error(err))
			return false, response.ServerError
		}
	}

	if len(tencentFiles) > 0 {
		_, err := s.fileRepo.DeleteFileByGroupIdAndName(ctx, tencentFiles, enum.FileStorageModeTencentCOS)
		if err != nil {
			logger.Log.Error("删除数据库中的腾讯云对象存储文件记录失败", zap.Error(err))
			return false, response.ServerError
		}
	}

	return true, nil
}

// pathHasFileGroup 根据路径字符串判断是否有文件组
// 如果路径为""、"/"、"." 则认为没有文件组
func (s *FileService) pathHasFileGroup(path string) bool {
	return len(path) > 0 && path != "" && path != "." && path != string(filepath.Separator)
}
