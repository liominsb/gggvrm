package controllers // Package controllers 控制器

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

//文章控制器

// ArticleListResponse 定义文章列表的轻量级返回结构（不包含 Content）
type ArticleListResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Preview      string    `json:"preview"`
	Likes        int       `json:"likes"`
	Views        int       `json:"views"`
	UserID       uint      `json:"user_id"`
	CategoryName string    `json:"category_name"` // 附加分类名
	Tags         []string  `json:"tags"`          // 附加标签数组
	CoverImg     string    `json:"cover_img"`     //【新增】封面图的 URL
	CreatedAt    time.Time `json:"created_at"`    // gorm.Model 自带的创建时间，列表通常需要展示
}

// ArticleCache 专门用来打包存入 Redis 的结构
type ArticleCache struct {
	Total int64                 `json:"total"`
	Data  []ArticleListResponse `json:"data"`
}

// 辅助函数：清理所有文章分页列表的缓存
func clearArticlesCache() {
	var cursor uint64 = 0
	var count int64 = 100
	var keys []string
	var err error
	for {
		keys, cursor, err = global.RedisDB.Scan(cursor, "articles:page:*", count).Result()
		if err != nil {
			log.Println(err)
			break
		}
		if len(keys) > 0 {
			err = global.RedisDB.Del(keys...).Err()
			if err != nil {
				log.Println(err)
			}
		}
		if cursor == 0 {
			break
		}
	}
}

// CreateArticle 创建文章
func CreateArticle(ctx *gin.Context) {
	var article models.Article

	var input struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Preview    string `json:"preview" binding:"required"`
		CategoryID uint   `json:"category_id"` // 新增分类 ID
		TagIDs     []uint `json:"tag_ids"`     // 新增标签 ID 数组
		CoverImg   string `json:"cover_img"`   //【新增】封面图的 URL
	}

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article = models.Article{
		Title:      input.Title,
		Content:    input.Content,
		Preview:    input.Preview,
		CategoryID: input.CategoryID,
		CoverImg:   input.CoverImg,
		UserID:     id.(uint),
	}

	if len(input.TagIDs) > 0 {
		var tags []models.Tag
		// 从数据库查出这些 ID 对应的真实标签数据
		if err := global.Db.Where("id IN ?", input.TagIDs).Find(&tags).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "查询标签失败"})
			return
		}
		// 将查出的标签对象切片赋值给 article
		article.Tags = tags
	}

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
	keyword := ctx.Query("keyword")           // 新增：获取搜索关键词参数
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

	if keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where("articles.title LIKE ? OR articles.preview LIKE ?", likePattern, likePattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "获取总数失败"})
		return
	}

	// 即使结果为空，也缓存（防止缓存穿透）
	if total == 0 {
		cacheObj := ArticleCache{Total: 0, Data: []ArticleListResponse{}}
		utils.Setcache(dynamicCacheKey, cacheObj) // 缓存空结果，设置较短过期时间
		ctx.JSON(http.StatusOK, gin.H{"data": []ArticleListResponse{}, "total": 0})
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

	//数据转换 (DTO 映射)
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
			Views:        a.Views,
			UserID:       a.UserID,
			CategoryName: a.Category.Name, // 需要预加载 Category
			Tags:         tagNames,
			CoverImg:     a.CoverImg, //【新增】封面图的 URL
			CreatedAt:    a.CreatedAt,
		})
	}

	cacheObj := ArticleCache{
		Total: total,
		Data:  response,
	}
	//将轻量级的数据存入缓存
	if err := utils.Setcache(dynamicCacheKey, cacheObj); err != nil {
		fmt.Printf("【Redis警告】缓存文章列表失败: %v\n", err)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize) //总页数，(21+5-1)/5=5
	//返回给前端
	ctx.JSON(http.StatusOK, gin.H{
		"data":        response,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetArticlesByID 根据ID获取文章
func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article
	var comments []models.Comment
	var likes string

	eg, gCtx := errgroup.WithContext(ctx.Request.Context())

	eg.Go(func() error {
		cacheKey := fmt.Sprintf("article:detail:%s", id)

		datastr, err := global.RedisDB.WithContext(gCtx).Get(cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(datastr), &article); err != nil {
				return err
			}
			return nil
		}
		if !errors.Is(err, redis.Nil) {
			return err
		}

		if err := global.Db.WithContext(gCtx).Preload("Category").Preload("Tags").Where("id = ?", id).First(&article).Error; err != nil {
			return err
		}

		if err := utils.Setcache(cacheKey, article); err != nil {
			fmt.Printf("【Redis警告】缓存文章详情失败: %v\n", err)
		}
		return nil
	})

	eg.Go(func() error {
		articleID, err := strconv.ParseUint(id, 10, 32)
		cacheKey := fmt.Sprintf("article:%d:comments", articleID)

		cacheData, err := global.RedisDB.WithContext(gCtx).Get(cacheKey).Result()

		if errors.Is(err, redis.Nil) {

			if err := global.Db.WithContext(gCtx).Where("article_id = ?", articleID).Find(&comments).Error; err != nil {
				return err
			}

			if err := utils.Setcache(cacheKey, comments); err != nil {
				fmt.Printf("【Redis警告】缓存文章评论失败: %v\n", err)
			}

			return nil
		}
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(cacheData), &comments); err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		var temp models.Article // 任务三专用的临时变量

		likeKey := "article:" + id + ":likes"

		likes, err = global.RedisDB.WithContext(gCtx).Get(likeKey).Result()

		if errors.Is(err, redis.Nil) {
			// 只查询 likes 字段，提高效率
			if err := global.Db.WithContext(gCtx).Select("likes").Where("id = ?", id).First(&temp).Error; err != nil {
				return err
			}

			if err := global.RedisDB.SetNX(likeKey, temp.Likes, 0).Err(); err != nil {
				return err
			}
			likes = strconv.Itoa(temp.Likes)
		} else if err != nil {
			return err
		}

		return nil
	})

	go func() {
		if err := global.RedisDB.Incr(fmt.Sprintf("article:%s:views", id)).Err(); err != nil {
			fmt.Printf("【Redis警告】增加文章浏览量失败: %v\n", err)
			return
		}
	}()

	if err := eg.Wait(); err != nil {
		// 精准识别是不是文章根本不存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "该文章不存在"})
			return
		}
		// 其他系统错误
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "数据组装失败", "detail": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"article":  article,
		"comments": comments,
		"likes":    likes,
	})
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

	//第一次删除缓存
	clearArticlesCache()
	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:comments", idA))
	global.RedisDB.Del(fmt.Sprintf("article:%s:likes", idA))

	if err := global.Db.Delete(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	time.Sleep(100 * time.Millisecond) //延时

	//第二次删除缓存
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
		CoverImg   string `json:"cover_img"`   //【新增】封面图的 URL
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
	//第一次删除缓存
	clearArticlesCache()
	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", articleID))

	updateData := map[string]interface{}{
		"title":       input.Title,
		"content":     input.Content,
		"preview":     input.Preview,
		"category_id": input.CategoryID,
		"cover_img":   input.CoverImg,
	}

	if err := global.Db.Model(&article).Preload("Category").Preload("Tags").Updates(updateData).Find(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} // 更新后重新查询一次，获取完整的文章数据（包括预加载的分类和标签）

	time.Sleep(100 * time.Millisecond) //延时

	// //第二次删除缓存（重要：文章修改后，详情缓存和列表分页缓存都会失效）
	clearArticlesCache()
	global.RedisDB.Del(fmt.Sprintf("article:detail:%s", articleID))

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功", "article": article})
}

// 获取ID小于cursor的limit个Article
// 请求：GET /articles?cursor=12345&limit=10
// 返回：最后一条的 ID 作为下次请求的 cursor
func GetArticlesByCursor(ctx *gin.Context) {
	cursorStr := ctx.DefaultQuery("cursor", "0")
	limitStr := ctx.DefaultQuery("limit", "10")

	cursor, _ := strconv.ParseUint(cursorStr, 10, 64)
	limit, _ := strconv.Atoi(limitStr)

	var articles []models.Article
	// 使用 WHERE id < cursor 代替 OFFSET
	global.Db.Where("id < ?", cursor).
		Order("id desc").
		Limit(limit + 1). // 多查一条判断是否有下一页
		Find(&articles)

	hasMore := len(articles) > limit
	if hasMore {
		articles = articles[:limit] // 去掉多查的那条
	}

	nextCursor := articles[len(articles)-1].ID

	ctx.JSON(http.StatusOK, gin.H{
		"data":        articles,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
