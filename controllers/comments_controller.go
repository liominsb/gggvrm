package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// CreateComments 创建评论
func CreateComment(ctx *gin.Context) {
	articleIDstring := ctx.Param("id")

	articleID, err := strconv.ParseUint(articleIDstring, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userid := id.(uint)

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	// 2. 解析前端传来的 application/json 数据
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空或格式错误"})
		return
	}

	comment := models.Comment{
		ArticleID: uint(articleID),
		UserID:    userid,
		Content:   input.Content,
	}

	if err := global.Db.Create(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cacheKey := fmt.Sprintf("article:%d:comments", articleID)
	if err := global.RedisDB.Del(cacheKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, comment)
}

func DelComment(ctx *gin.Context) {
	commentIDStr := ctx.Param("id")

	commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userid := id.(uint)

	comment := models.Comment{}
	if err := global.Db.Where("id = ?", commentID).First(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	article := models.Article{}
	if err := global.Db.Where("id = ?", comment.ArticleID).First(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userid != comment.UserID && userid != article.UserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := global.Db.Unscoped().Delete(&comment).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("article:%d:comments", article.ID)
	if err := global.RedisDB.Del(cacheKey).Err(); err != nil {
		fmt.Printf("【Redis警告】清理文章 %d 缓存失败: %v\n", comment.ArticleID, err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetComments 获取单个文章的所有评论
func GetComments(ctx *gin.Context) {

	articleIDstring := ctx.Param("id")
	articleID, err := strconv.ParseUint(articleIDstring, 10, 32)
	cacheKey := fmt.Sprintf("article:%d:comments", articleID)

	cacheData, err := global.RedisDB.Get(cacheKey).Result()

	if errors.Is(err, redis.Nil) {
		var comments []models.Comment

		if err := global.Db.Where("article_id = ?", articleID).Find(&comments).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := utils.Setcache(ctx, cacheKey, comments); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		ctx.JSON(http.StatusOK, comments)
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var comments []models.Comment
	if err := json.Unmarshal([]byte(cacheData), &comments); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, comments)
}
