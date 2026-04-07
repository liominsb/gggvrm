package controllers // Package controllers 控制器

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

//文章控制器

// ArticleListResponse 定义文章列表的轻量级返回结构（不包含 Content）
type ArticleListResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	Likes     int       `json:"likes"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"` // gorm.Model 自带的创建时间，列表通常需要展示
}

// 辅助函数：清理所有文章分页列表的缓存
func clearArticlesCache() {
	keys, err := global.RedisDB.Keys("articles:page:*").Result()
	if err == nil && len(keys) > 0 {
		global.RedisDB.Del(keys...)
	}
}

// CreateArticle 创建文章
func CreateArticle(ctx *gin.Context) {
	var article models.Article

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.UserID = id.(uint)

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clearArticlesCache()

	ctx.JSON(http.StatusCreated, article)
}

// GetArticles 获取所有文章
func GetArticles(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	dynamicCacheKey := fmt.Sprintf("articles:page:%d:size:%d", page, pageSize)

	cacheData, err := global.RedisDB.Get(dynamicCacheKey).Result()

	if err == nil {
		// 注意：现在缓存里存的已经是解析好的 ArticleListResponse 数组了
		var response []ArticleListResponse
		if err := json.Unmarshal([]byte(cacheData), &response); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "解析缓存失败"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data":      response,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}

	if !errors.Is(err, redis.Nil) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "读取缓存异常"})
		return
	}

	var articles []models.Article
	var total int64

	// 查总数
	global.Db.Model(&models.Article{}).Count(&total)

	// 分页计算
	offset := (page - 1) * pageSize

	// 核心优化：Omit("Content") 不查正文字段，极大节省内存和网络带宽
	if err := global.Db.Omit("Content").Order("id desc").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 数据转换 (DTO 映射)
	// 将查出来的 []models.Article 转换为干净的 []ArticleListResponse
	var response = make([]ArticleListResponse, 0) // 初始化为空切片，防止返回 null
	for _, a := range articles {
		response = append(response, ArticleListResponse{
			ID:        a.ID,
			Title:     a.Title,
			Preview:   a.Preview,
			Likes:     a.Likes,
			UserID:    a.UserID,
			CreatedAt: a.CreatedAt,
		})
	}

	// 5. 将轻量级的数据存入缓存
	utils.Setcache(ctx, dynamicCacheKey, response)

	// 6. 返回给前端
	ctx.JSON(http.StatusOK, gin.H{
		"data":      response,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetArticlesByID 根据ID获取文章
func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article

	cacheKey := fmt.Sprintf("article:detail:%s", id)

	datastr, err := global.RedisDB.Get(cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(datastr), &article); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, article)
		return
	}
	if !errors.Is(err, redis.Nil) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "读取缓存异常"})
		return
	}

	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	utils.Setcache(ctx, cacheKey, article)

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

	clearArticlesCache()

	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:comments", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:likes", idA))

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "删除成功"})
}
