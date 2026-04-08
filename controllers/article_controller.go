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
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Preview      string    `json:"preview"`
	Likes        int       `json:"likes"`
	UserID       uint      `json:"user_id"`
	CategoryName string    `json:"category_name"` // 附加分类名
	Tags         []string  `json:"tags"`          // 附加标签数组
	CreatedAt    time.Time `json:"created_at"`    // gorm.Model 自带的创建时间，列表通常需要展示
}

// ArticleCache 专门用来打包存入 Redis 的结构
type ArticleCache struct {
	Total int64                 `json:"total"`
	Data  []ArticleListResponse `json:"data"`
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
	categoryIDStr := ctx.Query("category_id") // 新增：获取分类筛选参数
	tagIDStr := ctx.Query("tag_id")           // 新增：获取标签筛选参数

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	categoryID, _ := strconv.Atoi(categoryIDStr)
	tagID, _ := strconv.Atoi(tagIDStr)

	if page <= 0 {
		page = 1
	}
	if page > 1000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "请求页码超出最大支持范围"})
		return
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	dynamicCacheKey := fmt.Sprintf("articles:page:%d:size:%d:cat:%d:tag:%d", page, pageSize, categoryID, tagID)

	cacheData, err := global.RedisDB.Get(dynamicCacheKey).Result()
	if err == nil {
		var cacheObj ArticleCache
		if err := json.Unmarshal([]byte(cacheData), &cacheObj); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "解析缓存失败"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data":      cacheObj.Data, // 从缓存对象中取出数据
			"total":     cacheObj.Total,
			"page":      page,
			"page_size": pageSize,
		})
		return
	}
	if !errors.Is(err, redis.Nil) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "读取缓存异常"})
		return
	}

	query := global.Db.Model(&models.Article{})

	// 1. 追加分类过滤条件 (一篇文章对一个分类，直接 Where)
	if categoryID > 0 {
		query = query.Where("articles.category_id = ?", categoryID)
	}

	// 2. 追加标签过滤条件 (多对多，需要联表查询 GORM 自动生成的 article_tags 中间表)
	if tagID > 0 {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tagID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取总数失败"})
		return
	}

	var articles []models.Article

	// 分页计算
	offset := (page - 1) * pageSize

	if int64(offset) < total {
		if err := query.Preload("Category").Preload("Tags").Omit("Content").Order("articles.id desc").Limit(pageSize).Offset(offset).Find(&articles).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 4. 数据转换 (DTO 映射)
	// 将查出来的 []models.Article 转换为干净的 []ArticleListResponse
	var response = make([]ArticleListResponse, 0) // 初始化为空切片，防止返回 null
	for _, a := range articles {
		var tagNames []string
		for _, t := range a.Tags {
			tagNames = append(tagNames, t.Name)
		}
		response = append(response, ArticleListResponse{
			ID:           a.ID,
			Title:        a.Title,
			Preview:      a.Preview,
			Likes:        a.Likes,
			UserID:       a.UserID,
			CategoryName: a.Category.Name, // 需要预加载 Category
			Tags:         tagNames,
			CreatedAt:    a.CreatedAt,
		})
	}

	cacheObj := ArticleCache{
		Total: total,
		Data:  response,
	}
	// 5. 将轻量级的数据存入缓存
	if err := utils.Setcache(ctx, dynamicCacheKey, cacheObj); err != nil {
		fmt.Printf("【Redis警告】缓存文章列表失败: %v\n", err)
	}

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

	if err := global.Db.Preload("Category").Preload("Tags").Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err := utils.Setcache(ctx, cacheKey, article); err != nil {
		fmt.Printf("【Redis警告】缓存文章详情失败: %v\n", err)
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

	clearArticlesCache()

	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:comments", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:likes", idA))

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "删除成功"})
}

// UpdateArticle 更新文章
func UpdateArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	var article models.Article
	// 1. 查找文章是否存在
	if err := global.Db.Where("id = ?", articleID).First(&article).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 2. 越权校验：只能修改自己的文章
	if article.UserID != userID.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "无权修改他人的文章"})
		return
	}

	// 3. 绑定新的输入数据 (利用一个临时结构体避免覆盖关键字段如 UserID)
	var input struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Preview    string `json:"preview" binding:"required"`
		CategoryID uint   `json:"category_id"` // 新增分类 ID
		TagIDs     []uint `json:"tag_ids"`     // 新增标签 ID 数组
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.TagIDs != nil {
		if len(input.TagIDs) > 0 {
			var tags []models.Tag
			// 查出这些 ID 对应的真实标签数据
			if err := global.Db.Where("id IN ?", input.TagIDs).Find(&tags).Error; err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "查询标签失败"})
				return
			}
			// 核心用法：使用 Replace 替换掉旧的关联。
			// GORM 会自动删除多余的旧标签关系，并插入新的关系。
			if err := global.Db.Model(&article).Association("Tags").Replace(tags); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签失败"})
				return
			}
		} else {
			// 如果前端传了一个空数组 []，说明用户想清空所有标签
			if err := global.Db.Model(&article).Association("Tags").Clear(); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "清空标签失败"})
				return
			}
		}
	}

	updateData := map[string]interface{}{
		"title":       input.Title,
		"content":     input.Content,
		"preview":     input.Preview,
		"category_id": input.CategoryID,
	}

	if err := global.Db.Model(&article).Preload("Category").Preload("Tags").Updates(updateData).Find(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} // 更新后重新查询一次，获取完整的文章数据（包括预加载的分类和标签）

	// 清理缓存（重要：文章修改后，详情缓存和列表分页缓存都会失效）
	clearArticlesCache()
	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", articleID))

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功", "article": article})
}
