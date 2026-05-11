package controllers

import (
	"gggvrm/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义一个 Controller 结构体，用于持有依赖
type LikeController struct {
	likeService service.LikeService
}

// 构造函数：在外部（比如 main.go 的路由配置里）初始化时传入依赖
func NewLikeController(likeService service.LikeService) *LikeController {
	return &LikeController{likeService: likeService}
}

// LikeArticle 赞成文章
func (c *LikeController) LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	likes, err := c.likeService.LikeArticle(ctx.Request.Context(), articleID)
	if err != nil {
		// 粗略根据错误信息判断状态码（严谨的做法是在 Service 返回自定义 Error 结构体包含 Code）
		if err.Error() == "文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "点赞成功: " + likes})
}

func (c *LikeController) GetArticlelikes(ctx *gin.Context) {
	articleID := ctx.Param("id")
	likes, err := c.likeService.GetArticleLikes(ctx.Request.Context(), articleID)
	if err != nil {
		// 粗略根据错误信息判断状态码（严谨的做法是在 Service 返回自定义 Error 结构体包含 Code）
		if err.Error() == "文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
