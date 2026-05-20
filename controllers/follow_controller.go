package controllers

import (
	"gggvrm/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FollowController 关注控制器，处理关注相关的 HTTP 请求
type FollowController struct {
	followService service.FollowService
}

// NewFollowController 创建关注控制器实例
func NewFollowController(followService service.FollowService) *FollowController {
	return &FollowController{followService: followService}
}

// ToggleFollow 关注/取消关注用户
func (c *FollowController) ToggleFollow(ctx *gin.Context) {
	followeeID := ctx.Param("id")
	userID, ex := ctx.Get("ID")
	if !ex {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	isFollowing, err := c.followService.ToggleFollow(ctx.Request.Context(), followeeID, userID.(uint))
	if err != nil {
		if err.Error() == "无效的用户ID" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "不能关注自己" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if err.Error() == "用户不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if isFollowing {
		ctx.JSON(http.StatusOK, gin.H{"message": "关注成功", "is_following": true})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "取消关注成功", "is_following": false})
	}
}

// GetFollowStatus 获取关注状态
func (c *FollowController) GetFollowStatus(ctx *gin.Context) {
	followeeID := ctx.Param("id")
	userID, ex := ctx.Get("ID")
	if !ex {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	isFollowing, err := c.followService.IsFollowing(ctx.Request.Context(), followeeID, userID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}

// GetFollowCounts 获取用户的关注数和粉丝数
func (c *FollowController) GetFollowCounts(ctx *gin.Context) {
	userID := ctx.Param("id")

	followingCount, followersCount, err := c.followService.GetFollowCounts(ctx.Request.Context(), userID)
	if err != nil {
		if err.Error() == "无效的用户ID" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"following_count": followingCount,
		"followers_count": followersCount,
	})
}

// GetFollowing 获取用户的关注列表
func (c *FollowController) GetFollowing(ctx *gin.Context) {
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

	users, total, err := c.followService.GetFollowing(ctx.Request.Context(), uint(targetID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	ctx.JSON(http.StatusOK, gin.H{
		"data":        users,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetFollowers 获取用户的粉丝列表
func (c *FollowController) GetFollowers(ctx *gin.Context) {
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

	users, total, err := c.followService.GetFollowers(ctx.Request.Context(), uint(targetID), page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	ctx.JSON(http.StatusOK, gin.H{
		"data":        users,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}
