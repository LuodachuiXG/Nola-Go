package admin

import (
	"nola-go/internal/middleware"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// DiaryAdminHandler 日记后端接口
type DiaryAdminHandler struct {
	diaryService *service.DiaryService
	tokenService *service.TokenService
}

func NewDiaryAdminHandler(dsv *service.DiaryService, tsv *service.TokenService) *DiaryAdminHandler {
	return &DiaryAdminHandler{
		diaryService: dsv,
		tokenService: tsv,
	}
}

func (h *DiaryAdminHandler) RegisterAdmin(r *gin.RouterGroup) {
	privateGroup := r.Group("/diary")
	privateGroup.Use(middleware.AuthMiddleware(h.tokenService))
	{
		// 添加日记
		privateGroup.POST("", h.addDiary)
		// 删除日记
		privateGroup.DELETE("", h.deleteDiaries)
		// 修改日记
		privateGroup.PUT("", h.updateDiary)
		// 获取日记
		privateGroup.GET("", h.getDiaries)
	}
}

// addDiary 添加日记
func (h *DiaryAdminHandler) addDiary(c *gin.Context) {
	var req *request.DiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.diaryService.AddDiary(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// deleteDiaries 删除日记
func (h *DiaryAdminHandler) deleteDiaries(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.ParamMismatch(c)
		return
	}

	if len(ids) == 0 {
		response.OkAndResponse(c, false)
		return
	}

	ret, err := h.diaryService.DeleteDiaries(c, ids)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// updateDiary 修改日记
func (h *DiaryAdminHandler) updateDiary(c *gin.Context) {
	var req *request.DiaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	if req.DiaryId == nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.diaryService.UpdateDiary(c, req)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}

// getDiaries 获取日记
func (h *DiaryAdminHandler) getDiaries(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)
	if err != nil {
		response.ParamMismatch(c)
		return
	}

	var req struct {
		Sort *enum.DiarySort `form:"sort"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamMismatch(c)
		return
	}

	ret, err := h.diaryService.DiariesPager(c, page, size, req.Sort)
	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}
	response.OkAndResponse(c, ret)
}
