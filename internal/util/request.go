package util

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// ShouldBindPager 绑定分页参数
// 如果 page 为 nil 或 0，page 和 size 都返回为 0。
func ShouldBindPager(c *gin.Context) (page, size int, err error) {
	var pager struct {
		Page *int `form:"page"`
		Size *int `form:"size"`
	}

	if err := c.ShouldBindQuery(&pager); err != nil {
		return 0, 0, errors.New("请求参数不匹配")
	}
	
	if pager.Page == nil {
		pager.Page = IntPtr(0)
	}

	if *pager.Page == 0 {
		// page 为 0，则 size 也为 0
		pager.Size = IntPtr(0)
	}

	if *pager.Page != 0 && *pager.Size == 0 {
		// 如果 page 不为 0，size 为 0，抛出参数不匹配异常
		return 0, 0, errors.New("请求参数不匹配")
	}

	if *pager.Page < 0 || *pager.Size < 0 {
		// 如果 page 或 size 小于 0，抛出参数不匹配异常
		return 0, 0, errors.New("请求参数不匹配")
	}

	return *pager.Page, *pager.Size, nil
}
