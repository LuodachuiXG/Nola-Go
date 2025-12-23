package api

import (
	"nola-go/internal/models/response"
	"nola-go/internal/service"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// DiaryApiHandler 日记博客接口
type DiaryApiHandler struct {
	diaryService *service.DiaryService
}

func NewDiaryApiHandler(diaryService *service.DiaryService) *DiaryApiHandler {
	return &DiaryApiHandler{
		diaryService: diaryService,
	}
}

// RegisterApi 注册日记博客路由
func (h *DiaryApiHandler) RegisterApi(r *gin.RouterGroup) {
	publicGroup := r.Group("/diary")
	{
		publicGroup.GET("", h.getDiaries)
	}
}

// getDiaries 获取日记
func (h *DiaryApiHandler) getDiaries(c *gin.Context) {
	page, size, err := util.ShouldBindPager(c)

	if err != nil {
		response.FailAndResponse(c, err.Error())
	}

	pager, err := h.diaryService.DiariesPager(c, page, size, nil)

	if err != nil {
		response.FailAndResponse(c, err.Error())
		return
	}

	response.OkAndResponse(c, pager)
}
