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
	comment := models.Comment{
		ArticleID: uint(articleID),
		UserID:    id.(uint),
		Content:   ctx.PostForm("content"),
	}
	if err := global.Db.AutoMigrate(&comment); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

		utils.Setcache(ctx, cacheKey, comments)
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
