package controllers

import (
	"gggvrm/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FeedController 个人 Feed 流控制器
type FeedController struct {
	feedService service.FeedService
}

// NewFeedController 构造函数
func NewFeedController(feedService service.FeedService) *FeedController {
	return &FeedController{feedService: feedService}
}

// GetUserFeed 获取当前登录用户的个人 Feed 流
// GET /articles/feed?page=1&page_size=10
func (c *FeedController) GetUserFeed(ctx *gin.Context) {
	// 从 JWT 中间件获取当前登录用户 ID
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	data, total, err := c.feedService.GetUserFeed(ctx.Request.Context(), userID.(uint), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	ctx.JSON(http.StatusOK, gin.H{
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
