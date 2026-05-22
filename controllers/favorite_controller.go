package controllers

import (
	"gggvrm/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FavoriteController 收藏控制器，处理收藏相关的 HTTP 请求
type FavoriteController struct {
	favoriteService service.FavoriteService
}

// NewFavoriteController 创建收藏控制器实例
func NewFavoriteController(favoriteService service.FavoriteService) *FavoriteController {
	return &FavoriteController{favoriteService: favoriteService}
}

// ToggleFavorite 收藏/取消收藏文章
func (c *FavoriteController) ToggleFavorite(ctx *gin.Context) {
	articleID := ctx.Param("id")
	userID, ex := ctx.Get("ID")
	if !ex {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	isFav, err := c.favoriteService.ToggleFavorite(ctx.Request.Context(), articleID, userID.(uint))
	if err != nil {
		if err.Error() == "无效的文章ID" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if isFav {
		ctx.JSON(http.StatusOK, gin.H{"message": "收藏成功", "is_favorited": true})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "取消收藏成功", "is_favorited": false})
	}
}

// GetFavoriteStatus 获取文章收藏状态
func (c *FavoriteController) GetFavoriteStatus(ctx *gin.Context) {
	articleID := ctx.Param("id")
	userID, ex := ctx.Get("ID")
	if !ex {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	isFav, err := c.favoriteService.IsFavorited(ctx.Request.Context(), articleID, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"is_favorited": isFav})
}

// GetFavoriteCount 获取文章收藏数
func (c *FavoriteController) GetFavoriteCount(ctx *gin.Context) {
	articleID := ctx.Param("id")

	count, err := c.favoriteService.GetFavoriteCount(ctx.Request.Context(), articleID)
	if err != nil {
		if err.Error() == "无效的文章ID" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"favorites": count})
}

// GetUserFavorites 获取当前登录用户的收藏列表
func (c *FavoriteController) GetUserFavorites(ctx *gin.Context) {
	userID, ex := ctx.Get("ID")
	if !ex {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	articles, total, err := c.favoriteService.GetUserFavorites(ctx.Request.Context(), userID.(uint), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	ctx.JSON(http.StatusOK, gin.H{
		"data":        articles,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetUserFavoritesById 根据用户ID获取收藏列表（可查看其他用户的收藏）
func (c *FavoriteController) GetUserFavoritesById(ctx *gin.Context) {
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil || targetID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	articles, total, err := c.favoriteService.GetUserFavorites(ctx.Request.Context(), uint(targetID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	ctx.JSON(http.StatusOK, gin.H{
		"data":        articles,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
