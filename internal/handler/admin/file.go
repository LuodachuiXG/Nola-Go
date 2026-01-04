package admin

import (
	"fmt"
	"mime/multipart"
	"nola-go/internal/file/config"
	"nola-go/internal/middleware"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FileAdminHandler 文件后端路由接口
type FileAdminHandler struct {
	fileService  *service.FileService
	tokenService *service.TokenService
}

func NewFileAdminHandler(fsv *service.FileService, tsv *service.TokenService) *FileAdminHandler {
	return &FileAdminHandler{
		fileService:  fsv,
		tokenService: tsv,
	}
}

// RegisterAdmin 注册文件后端路由
func (h *FileAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	privateGroup := r.Group("/file")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))

	// 文件存储方式相关路由
	fileStorageModeRouting := privateGroup.Group("/mode")
	{
		// 获取已经设置的所有存储方式
		fileStorageModeRouting.GET("", h.getModes)
		// 设置腾讯云对象存储
		fileStorageModeRouting.POST("/tencent_cos", h.setTencentCOS)
		// 删除腾讯云对象存储配置
		fileStorageModeRouting.DELETE("/tencent_cos", h.deleteTencentCOS)
		// 获取腾讯云对象存储配置
		fileStorageModeRouting.GET("/tencent_cos", h.getTencentCOS)
	}

	// 文件相关路由
	{
		// 添加文件
		privateGroup.POST("", h.addFile)
		// 添加文件记录
		privateGroup.POST("/record", h.addFileRecord)
		// 根据文件 ID 数组删除文件
		privateGroup.DELETE("", h.deleteFilesByIds)
		// 根据文件索引数组删除文件
		privateGroup.DELETE("/name", h.deleteFilesByNameIndexes)
		// 移动文件
		privateGroup.PUT("", h.moveFiles)
		// 获取文件
		privateGroup.GET("", h.getFiles)
	}

	// 文件组相关路由
	fileGroupRouting := privateGroup.Group("/group")
	{
		// 添加文件组
		fileGroupRouting.POST("", h.addFileGroup)
		// 删除文件组
		fileGroupRouting.DELETE("/:fileGroupId", h.deleteFileGroup)
		// 修改文件组
		fileGroupRouting.PUT("", h.updateFileGroup)
		// 根据文件组存储方式获取文件组
		fileGroupRouting.GET("", h.getFileGroupByMode)
	}

}

// getModes 获取已经设置的所有存储方式
func (h *FileAdminHandler) getModes(c *gin.Context) {
	ret, err := h.fileService.GetModes(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
	}
	response.OkAndResponse(c, ret)
}

// setTencentCOS 设置腾讯云对象存储
func (h *FileAdminHandler) setTencentCOS(c *gin.Context) {
	var req *config.TencentCosConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println(fmt.Sprintf("%#v", req))
		response.ParamMismatch(c)
		return
	}

	ret, err := h.fileService.SetTencentCOS(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteTencentCOS 删除腾讯云对象存储
func (h *FileAdminHandler) deleteTencentCOS(c *gin.Context) {
	ret, err := h.fileService.DeleteTencentCOS(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getTencentCOS 获取腾讯云对象存储配置
func (h *FileAdminHandler) getTencentCOS(c *gin.Context) {
	ret, err := h.fileService.GetTencentCOS(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// addFile 添加文件
func (h *FileAdminHandler) addFile(c *gin.Context) {
	var req struct {
		File    *multipart.FileHeader `form:"file" binding:"required"`
		GroupId *uint                 `form:"fileGroupId"`
		Mode    *string               `form:"storageMode"`
	}

	if err := c.ShouldBind(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.Mode != nil && enum.FileStorageModeValueOf(*req.Mode) == nil {
		response.ParamMismatch(c)
		return
	}

	// 文件为 nil，不执行任何操作
	if req.File == nil {
		response.OkAndResponse(c, false)
		return
	}

	// 文件名
	originName := req.File.Filename
	if util.StringIsBlank(originName) {
		originName = "未命名文件"
	}

	// 默认存储策略
	defaultMode := enum.FileStorageModeLocal
	if req.Mode != nil {
		defaultMode = *enum.FileStorageModeValueOf(*req.Mode)
	}

	// 打开文件流
	file, err := req.File.Open()
	if err != nil {
		response.FailAndResponse(c, "打开文件失败，请检查服务器日志")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	ret, err := h.fileService.UploadFile(
		c,
		file,
		originName,
		// 默认本地存储
		defaultMode,
		req.GroupId,
		&req.File.Size,
	)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, ret)
}

// addFileRecord 添加文件记录
func (h *FileAdminHandler) addFileRecord(c *gin.Context) {
	var req *request.FileRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.fileService.UploadFileRecord(c, *req)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteFilesByIds 根据文件 ID 数组删除文件
func (h *FileAdminHandler) deleteFilesByIds(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.fileService.DeleteFiles(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteFilesByNameIndexes 根据文件索引数组删除文件
func (h *FileAdminHandler) deleteFilesByNameIndexes(c *gin.Context) {
	var req []*models.FileIndex
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.fileService.DeleteFilesByFileIndexes(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// moveFiles 移动文件
func (h *FileAdminHandler) moveFiles(c *gin.Context) {
	var req *request.FileMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.fileService.MoveFiles(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getFiles 获取文件
func (h *FileAdminHandler) getFiles(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.FailAndResponse(c, err.Error())
	}

	var req struct {
		Sort    *string `form:"sort"`
		Mode    *string `form:"mode"`
		GroupId *string `form:"groupId"`
		Key     *string `form:"key"`
	}

	var sortEnum *enum.FileSort
	var modeEnum *enum.FileStorageMode

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.Sort != nil {
		sortEnum = enum.FileSortValueOf(*req.Sort)
		if sortEnum == nil {
			response.ParamMismatch(c)
			return
		}
	}

	if req.Mode != nil {
		modeEnum = enum.FileStorageModeValueOf(*req.Mode)
		if modeEnum == nil {
			response.ParamMismatch(c)
			return
		}
	}

	// 检查文件组 ID 是否合法
	var groupId *uint
	if req.GroupId != nil {
		if groupIdUint, err := strconv.ParseUint(*req.GroupId, 10, 32); err == nil {
			groupId = new(uint)
			*groupId = uint(groupIdUint)
		} else {
			response.ParamMismatch(c)
			return
		}
	}

	ret, err := h.fileService.GetFiles(c, page, size, sortEnum, modeEnum, groupId, req.Key)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// addFileGroup 添加文件组
func (h *FileAdminHandler) addFileGroup(c *gin.Context) {
	var req *request.FileGroupAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.fileService.AddFileGroup(c, *req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteFileGroup 删除文件组
func (h *FileAdminHandler) deleteFileGroup(c *gin.Context) {
	var req struct {
		FileGroupId uint `uri:"fileGroupId"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.fileService.DeleteFileGroup(c, req.FileGroupId)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateFileGroup 修改文件组
func (h *FileAdminHandler) updateFileGroup(c *gin.Context) {
	var req *request.FileGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}
	ret, err := h.fileService.UpdateFileGroup(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getFileGroupByMode 根据文件组存储方式获取文件组
func (h *FileAdminHandler) getFileGroupByMode(c *gin.Context) {
	var req struct {
		FileStorageMode *string `form:"fileStorageMode"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	var modeEnum *enum.FileStorageMode = nil
	if req.FileStorageMode != nil {
		modeEnum = enum.FileStorageModeValueOf(*req.FileStorageMode)
		if modeEnum == nil {
			response.ParamMismatch(c)
			return
		}
	}

	ret, err := h.fileService.GetFileGroupsByMode(c, modeEnum)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
