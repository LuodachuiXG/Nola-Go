package repository

import (
	"context"
	"errors"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/util"

	"gorm.io/gorm"
)

type FileRepository interface {
	// SetFileStorageConfig 设置文件存储方式配置
	SetFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode, config string) (bool, error)
	// DeleteFileStorageConfig 删除文件存储方式
	DeleteFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode) (bool, error)
	// GetFileStorageConfig 获取文件存储方式配置
	GetFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode) (*string, error)
	// GetFileCountByMode 获取指定的存储方式的文件数量
	GetFileCountByMode(ctx context.Context, storageMode enum.FileStorageMode) (int64, error)
	// GetFileCountByGroup 获取指定的文件组下的文件数量
	GetFileCountByGroup(ctx context.Context, groupId uint) (int64, error)
	// GetFileCount 获取总文件数量
	GetFileCount(ctx context.Context) (int64, error)
	// GetModes 获取所有已经设置过的存储策略
	GetModes(ctx context.Context) ([]enum.FileStorageMode, error)
	// AddFileGroup 添加文件组
	AddFileGroup(ctx context.Context, fileGroup request.FileGroupAddRequest) (*models.FileGroup, error)
	// DeleteFileGroup 删除文件组
	DeleteFileGroup(ctx context.Context, groupId uint) (bool, error)
	// UpdateFileGroup 修改文件组
	UpdateFileGroup(ctx context.Context, fileGroup request.FileGroupUpdateRequest) (bool, error)
	// GetFileGroupById 根据文件组 ID 获取文件组
	GetFileGroupById(ctx context.Context, groupId uint) (*models.FileGroup, error)
	// GetFileGroupsByPath 根据文件组路径获取文件组
	GetFileGroupsByPath(ctx context.Context, paths []string) ([]*models.FileGroup, error)
	// GetFileGroupsByMode 根据文件存储方式获取文件组
	GetFileGroupsByMode(ctx context.Context, storageMode *enum.FileStorageMode) ([]*models.FileGroup, error)
	// AddFile 添加文件
	AddFile(ctx context.Context, file models.File) (*models.File, error)
	// DeleteFileById 根据文件 ID 删除文件
	DeleteFileById(ctx context.Context, fileId uint) (bool, error)
	// DeleteFileByIds 根据文件 ID 数组批量删除文件
	DeleteFileByIds(ctx context.Context, ids []uint) (bool, error)
	// DeleteFileByGroupIdAndName 根据文件组 ID 和文件名批量删除文件
	DeleteFileByGroupIdAndName(ctx context.Context, pairs []*models.Pair[*uint, string], storageMode enum.FileStorageMode) (bool, error)
	// UpdateFile 修改文件
	UpdateFile(ctx context.Context, file models.File) (bool, error)
	// UpdateFiles 批量修改文件
	UpdateFiles(ctx context.Context, files []models.File) (bool, error)
	// GetFile 根据文件名、文件组 ID 和文件存储方式获取文件
	GetFile(ctx context.Context, fileName string, groupId *uint, storageMode enum.FileStorageMode) (*models.File, error)
	// GetFileWithGroupByIds 根据文件 ID 数组，获取文件和文件组数据类
	GetFileWithGroupByIds(ctx context.Context, ids []uint) ([]*models.FileWithGroup, error)
	// GetFileWithGroups 获取文件
	GetFileWithGroups(ctx context.Context, page, size int, sort *enum.FileSort, mode *enum.FileStorageMode, groupId *uint, key *string) (*models.Pager[models.FileWithGroup], error)
	// GetFileByIds 根据文件 ID 数组批量获取所有文件
	GetFileByIds(ctx context.Context, ids []uint) ([]*models.File, error)
}

type fileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) FileRepository {
	return &fileRepo{
		db: db,
	}
}

// SetFileStorageConfig 设置文件存储方式配置
//   - storageMode: 文件存储方式
//   - config: 配置
func (r *fileRepo) SetFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode, config string) (bool, error) {
	strConfig, err := r.GetFileStorageConfig(ctx, storageMode)
	if err != nil {
		return false, err
	}

	if strConfig == nil {
		// 新增配置
		add := &models.FileStorageModes{
			StorageMode: &storageMode,
			Config:      config,
		}

		err := r.db.WithContext(ctx).Create(add).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// 修改配置
	update := map[string]any{
		"config": config,
	}
	err = r.db.WithContext(ctx).
		Model(&models.FileStorageModes{}).
		Where("storage_mode = ?", storageMode).
		Updates(update).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteFileStorageConfig 删除文件存储方式
//   - storageMode: 文件存储方式
func (r *fileRepo) DeleteFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode) (bool, error) {
	tx := r.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除对应的文件组
	err := tx.Where("storage_mode = ?", storageMode).Delete(&models.FileGroup{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除文件存储方式
	ret := tx.Where("storage_mode = ?", storageMode).Delete(&models.FileStorageModes{})
	if ret.Error != nil {
		tx.Rollback()
		return false, ret.Error
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// GetFileStorageConfig 获取文件存储方式配置
func (r *fileRepo) GetFileStorageConfig(ctx context.Context, storageMode enum.FileStorageMode) (*string, error) {
	var mode *models.FileStorageModes
	err := r.db.WithContext(ctx).
		Model(&models.FileStorageModes{}).
		Where("storage_mode = ?", storageMode).
		First(&mode).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || mode == nil {
			return nil, nil
		}
		return nil, err
	}
	return &mode.Config, nil
}

// GetFileCountByMode 获取指定的存储方式的文件数量
func (r *fileRepo) GetFileCountByMode(ctx context.Context, storageMode enum.FileStorageMode) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("storage_mode = ?", storageMode).
		Count(&count).Error
	return count, err
}

// GetFileCountByGroup 获取指定的文件组下的文件数量
func (r *fileRepo) GetFileCountByGroup(ctx context.Context, groupId uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("file_group_id = ?", groupId).
		Count(&count).Error
	return count, err
}

// GetFileCount 获取总文件数量
func (r *fileRepo) GetFileCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Count(&count).Error
	return count, err
}

// GetModes 获取所有已经设置过的存储策略
func (r *fileRepo) GetModes(ctx context.Context) ([]enum.FileStorageMode, error) {
	var ret []enum.FileStorageMode
	err := r.db.WithContext(ctx).
		Model(&models.FileStorageModes{}).
		Pluck("storage_mode", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// AddFileGroup 添加文件组
func (r *fileRepo) AddFileGroup(ctx context.Context, fileGroup request.FileGroupAddRequest) (*models.FileGroup, error) {
	add := &models.FileGroup{
		DisplayName: fileGroup.DisplayName,
		Path:        fileGroup.Path,
		StorageMode: fileGroup.StorageMode,
	}

	err := r.db.WithContext(ctx).Create(add).Error
	if err != nil {
		return nil, err
	}
	return add, nil
}

// DeleteFileGroup 删除文件组
func (r *fileRepo) DeleteFileGroup(ctx context.Context, groupId uint) (bool, error) {
	ret := r.db.WithContext(ctx).
		Where("file_group_id = ?", groupId).
		Delete(&models.FileGroup{})

	if ret.Error != nil {
		return false, ret.Error
	}

	return ret.RowsAffected > 0, nil
}

// UpdateFileGroup 修改文件组，文件存储方式不能修改
func (r *fileRepo) UpdateFileGroup(ctx context.Context, fileGroup request.FileGroupUpdateRequest) (bool, error) {
	updates := map[string]any{
		"display_name": fileGroup.DisplayName,
	}

	ret := r.db.WithContext(ctx).
		Model(&models.FileGroup{}).
		Where("file_group_id = ?", fileGroup.FileGroupId).
		Updates(updates)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// GetFileGroupById 根据文件组 ID 获取文件组
func (r *fileRepo) GetFileGroupById(ctx context.Context, groupId uint) (*models.FileGroup, error) {
	var fileGroup *models.FileGroup
	err := r.db.WithContext(ctx).
		Model(&models.FileGroup{}).
		Where("file_group_id = ?", groupId).
		First(&fileGroup).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return fileGroup, nil
}

// GetFileGroupsByMode 根据文件存储方式获取文件组
func (r *fileRepo) GetFileGroupsByMode(ctx context.Context, storageMode *enum.FileStorageMode) ([]*models.FileGroup, error) {
	var ret []*models.FileGroup
	base := r.db.WithContext(ctx).Model(&models.FileGroup{})
	if storageMode != nil {
		base = base.Where("storage_mode = ?", storageMode)
	}

	err := base.Find(&ret).Error

	if err != nil {
		return nil, err
	}
	return ret, nil
}

// GetFileGroupsByPath 根据文件组路径获取文件组
func (r *fileRepo) GetFileGroupsByPath(ctx context.Context, paths []string) ([]*models.FileGroup, error) {
	var ret []*models.FileGroup
	err := r.db.WithContext(ctx).
		Model(&models.FileGroup{}).
		Where("path IN ?", paths).
		Find(&ret).Error

	if err != nil {
		return []*models.FileGroup{}, err
	}
	return ret, nil
}

// AddFile 添加文件
func (r *fileRepo) AddFile(ctx context.Context, file models.File) (*models.File, error) {
	add := &models.File{
		DisplayName: file.DisplayName,
		Size:        file.Size,
		StorageMode: file.StorageMode,
		CreateTime:  file.CreateTime,
	}

	if file.FileGroupId != nil {
		add.FileGroupId = file.FileGroupId
	}

	err := r.db.WithContext(ctx).Create(add).Error

	if err != nil {
		return nil, err
	}
	return add, nil
}

// DeleteFileById 根据文件 ID 删除文件
func (r *fileRepo) DeleteFileById(ctx context.Context, fileId uint) (bool, error) {
	ret := r.db.WithContext(ctx).
		Where("file_id = ?", fileId).
		Delete(&models.File{})
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// DeleteFileByIds 根据文件 ID 数组批量删除文件
func (r *fileRepo) DeleteFileByIds(ctx context.Context, ids []uint) (bool, error) {
	ret := r.db.WithContext(ctx).
		Where("file_id IN ?", ids).
		Delete(&models.File{})
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// DeleteFileByGroupIdAndName 根据文件组 ID 和文件名批量删除文件
//   - ctx: 上下文
//   - pairs: 文件组 ID 和文件名对数组
//   - storageMode: 文件存储方式
func (r *fileRepo) DeleteFileByGroupIdAndName(ctx context.Context, pairs []*models.Pair[*uint, string], storageMode enum.FileStorageMode) (bool, error) {
	// 每批最多处理 500 条，防止拼接的参数过多
	const batchSize = 500

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 分开处理有文件组和没有文件组的数据
	var withGroup []*models.Pair[*uint, string]
	var onlyName []string
	for _, pair := range pairs {
		if pair.First != nil {
			withGroup = append(withGroup, pair)
		} else {
			onlyName = append(onlyName, pair.Second)
		}
	}

	// 影响行数
	var deleteCount int64 = 0

	// 先删除有文件组的记录
	if len(withGroup) > 0 {
		for _, chunk := range util.Chunk(withGroup, batchSize) {
			var args [][]any
			for _, p := range chunk {
				args = append(args, []any{*p.First, p.Second})
			}

			ret := tx.Where("storage_mode = ? AND (file_group_id, display_name) IN ?", storageMode, args).
				Delete(&models.File{})

			if ret.Error != nil {
				tx.Rollback()
				return false, ret.Error
			}

			deleteCount += ret.RowsAffected
		}
	}

	// 删除没有文件组的记录
	if len(onlyName) > 0 {
		for _, chunk := range util.Chunk(onlyName, batchSize) {
			ret := tx.Where("storage_mode = ? AND file_group_id IS NULL AND display_name IN ?", storageMode, chunk).
				Delete(&models.File{})

			if ret.Error != nil {
				tx.Rollback()
				return false, ret.Error
			}

			deleteCount += ret.RowsAffected
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return deleteCount > 0, nil
}

// UpdateFile 修改文件
func (r *fileRepo) UpdateFile(ctx context.Context, file models.File) (bool, error) {
	update := map[string]any{
		"file_group_id": file.FileGroupId,
		"display_name":  file.DisplayName,
		"size":          file.Size,
		"storage_mode":  file.StorageMode,
	}
	ret := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("file_id = ?", file.FileId).
		Updates(update)
	if ret.Error != nil {
		return false, ret.Error
	}
	return ret.RowsAffected > 0, nil
}

// UpdateFiles 批量修改文件
func (r *fileRepo) UpdateFiles(ctx context.Context, files []models.File) (bool, error) {

	// 成功条数
	count := 0

	tx := r.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 开始更新
	for _, file := range files {
		update := map[string]any{
			"file_group_id": file.FileGroupId,
			"display_name":  file.DisplayName,
			"size":          file.Size,
			"storage_mode":  file.StorageMode,
		}

		ret := tx.Model(&models.File{}).
			Where("file_id = ?", file.FileId).
			Updates(update)

		if ret.Error != nil {
			tx.Rollback()
			return false, ret.Error
		}

		if ret.RowsAffected > 0 {
			count++
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFile 根据文件名、文件组 ID 和文件存储方式获取文件
func (r *fileRepo) GetFile(ctx context.Context, fileName string, groupId *uint, storageMode enum.FileStorageMode) (*models.File, error) {

	var file *models.File

	baseQuery := r.db.WithContext(ctx).Model(&models.File{})

	if groupId != nil {
		baseQuery = baseQuery.Where("file_group_id = ?", groupId)
	}

	baseQuery = baseQuery.Where("display_name = ? AND storage_mode = ?", fileName, storageMode)

	err := baseQuery.First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return file, nil
}

// GetFileWithGroupByIds 根据文件 ID 数组，获取文件和文件组数据类
func (r *fileRepo) GetFileWithGroupByIds(ctx context.Context, ids []uint) ([]*models.FileWithGroup, error) {
	var ret []*models.FileWithGroup
	err := r.db.WithContext(ctx).
		Table("file f").
		Select("f.file_id as fileId, f.file_group_id as fileGroupId, f.display_name as fileName, fg.display_name as fileGroupName, fg.path as fileGroupPath, f.size as size, f.storage_mode as storageMode, f.create_time as createTime").
		Joins("LEFT JOIN file_group fg ON f.file_group_id = fg.file_group_id").
		Where("f.file_id IN ?", ids).
		Scan(&ret).Error

	if err != nil {
		return nil, err
	}

	return ret, nil
}

// GetFileWithGroups 获取文件
//   - page: 页码
//   - size: 每页条数
//   - sort: 排序方式
//   - mode: 文件存储方式
//   - groupId: 文件组 ID
//   - key: 关键字
func (r *fileRepo) GetFileWithGroups(
	ctx context.Context,
	page, size int,
	sort *enum.FileSort,
	mode *enum.FileStorageMode,
	groupId *uint,
	key *string,
) (*models.Pager[models.FileWithGroup], error) {
	baseQuery := r.db.WithContext(ctx).
		Table("file f").
		Select("f.file_id as fileId, f.file_group_id as fileGroupId, f.display_name as fileName, fg.display_name as fileGroupName, fg.path as fileGroupPath, f.size as size, f.storage_mode as storageMode, f.create_time as createTime").
		Joins("LEFT JOIN file_group fg ON f.file_group_id = fg.file_group_id")

	if mode != nil {
		baseQuery = baseQuery.Where("f.storage_mode = ?", mode)
	}

	if groupId != nil {
		baseQuery = baseQuery.Where("f.file_group_id = ?", groupId)
	}

	if !util.StringIsNilOrBlank(key) {
		baseQuery = baseQuery.Where("f.display_name LIKE ?", "%"+*key+"%")
	}

	if sort != nil {
		switch *sort {
		case enum.FileSortCreateTimeDesc:
			baseQuery = baseQuery.Order("f.create_time DESC")
		case enum.FileSortCreateTimeAsc:
			baseQuery = baseQuery.Order("f.create_time ASC")
		case enum.FileSortSizeDesc:
			baseQuery = baseQuery.Order("f.size DESC")
		case enum.FileSortSizeAsc:
			baseQuery = baseQuery.Order("f.size ASC")
		}
	}

	if page == 0 {
		// 获取所有文件
		var ret []*models.FileWithGroup
		err := baseQuery.Scan(&ret).Error
		if err != nil {
			return nil, err
		}
		return &models.Pager[models.FileWithGroup]{
			Page:       0,
			Size:       0,
			Data:       ret,
			TotalData:  int64(len(ret)),
			TotalPages: 1,
		}, nil
	}

	// page 不等于 0，分页查询
	pager, err := db.PagerBuilder[models.FileWithGroup](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return baseQuery
	})

	if err != nil {
		return nil, err
	}
	return pager, nil
}

// GetFileByIds 根据文件 ID 数组批量获取所有文件
func (r *fileRepo) GetFileByIds(ctx context.Context, ids []uint) ([]*models.File, error) {
	var ret []*models.File
	err := r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("file_id IN ?", ids).
		Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
