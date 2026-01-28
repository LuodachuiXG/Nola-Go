package admin

import (
	"fmt"
	"io"
	"mime/multipart"
	"nola-go/internal/logger"
	"nola-go/internal/middleware"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// BackupAdminHandler 备份后端接口
type BackupAdminHandler struct {
	postService  *service.PostService
	tokenService *service.TokenService
}

func NewBackupAdminHandler(psv *service.PostService, tsc *service.TokenService) *BackupAdminHandler {
	return &BackupAdminHandler{
		postService:  psv,
		tokenService: tsc,
	}
}

// RegisterAdmin 注册备份后端接口
func (h *BackupAdminHandler) RegisterAdmin(r *gin.RouterGroup) {

	privateGroup := r.Group("/backup")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 导入文章
		privateGroup.POST("/post", h.importPost)
		// 导出文章
		privateGroup.GET("/post", h.exportPost)
	}
}

// importPost 导入文章
// 可以上传 Markdown 或 PlainText 文章，文件名作为文章名。
func (h *BackupAdminHandler) importPost(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	// 获取 key 为 files 的所有文件
	files := form.File["files"]
	if len(files) == 0 {
		response.ParamMismatch(c)
		return
	}

	// 允许的文件后缀
	allowedExts := map[string]bool{
		".md":  true,
		".txt": true,
	}

	// 待添加文件列表
	var fileList []*multipart.FileHeader
	// 添加失败的消息列表
	var errorResult []string

	// 检验文件
	for _, file := range files {
		// 获取后缀名
		name := filepath.Base(file.Filename)
		ext := strings.ToLower(filepath.Ext(name))

		if _, ok := allowedExts[ext]; !ok {
			errorResult = append(errorResult, fmt.Sprintf("%s，文件类型错误", name))
		} else {
			fileList = append(fileList, file)
		}
	}

	// 读取文章内容
	var fileNameList []string
	var fileContentList []string
	for _, file := range fileList {
		f, err := file.Open()
		if err != nil {
			errorResult = append(errorResult, fmt.Sprintf("%s，读取文件失败", file.Filename))
			logger.Log.Error(fmt.Sprintf("读取文件失败：%s", err))
		} else {
			// 不包含后缀名的文件名
			name := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
			content, err := io.ReadAll(f)
			_ = f.Close()
			if err != nil {
				errorResult = append(errorResult, fmt.Sprintf("%s，读取文件失败", file.Filename))
				logger.Log.Error(fmt.Sprintf("读取文件失败：%s", err))
				continue
			}

			fileNameList = append(fileNameList, name)
			fileContentList = append(fileContentList, string(content))
		}
	}

	// 添加文章
	ret, err := h.postService.AddPostByNamesAndContents(c, fileNameList, fileContentList)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, map[string]any{
		// 文章总数
		"fileCount": len(fileNameList),
		// 成功数量
		"successCount": len(ret),
		// 失败数量
		"failCount": len(errorResult),
		// 失败信息
		"errorResult": errorResult,
	})
}

// exportPost 导出文章
func (h *BackupAdminHandler) exportPost(c *gin.Context) {
	ret, err := h.postService.ExportPosts(c)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
