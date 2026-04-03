package controllers // Package controllers 控制器

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

//文章控制器

var cacheKey = "articles"

// CreateArticle 创建文章
func CreateArticle(ctx *gin.Context) {
	var article models.Article

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	if err := ctx.ShouldBind(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.UserID = id.(uint)

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := global.RedisDB.Del(cacheKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

// GetArticles 获取所有文章
func GetArticles(ctx *gin.Context) {

	cacheData, err := global.RedisDB.Get(cacheKey).Result()

	if errors.Is(err, redis.Nil) {
		var articles []models.Article

		if err := global.Db.Find(&articles).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		utils.Setcache(ctx, cacheKey, articles)

		ctx.JSON(http.StatusOK, articles)
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var articles []models.Article
	if err := json.Unmarshal([]byte(cacheData), &articles); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, articles)
}

// GetArticlesByID 根据ID获取文章
func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article

	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, article)
}

func DelArticle(ctx *gin.Context) {
	idA := ctx.Param("id")
	idU, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	var article models.Article

	if err := global.Db.Where("id = ?", idA).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if article.UserID != idU.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权限删除该文章"})
		return
	}

	if err := global.Db.Delete(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := global.RedisDB.Del(cacheKey).Err(); err != nil {
		fmt.Printf("【警告】文章已删除，但清理文章列表缓存失败: %v\n", err)
	}

	global.RedisDB.Del(fmt.Sprintf("article:%s:comments", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:likes", idA))

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
