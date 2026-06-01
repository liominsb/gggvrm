package controllers // Package controllers 控制器

import (
	"bytes"
	"context"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/rag_grpc"
	"gggvrm/service"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

//文章控制器

type ArticleController struct {
	articleService service.ArticleService
}

// 构造函数
func NewArticleController(articleService service.ArticleService) *ArticleController {
	return &ArticleController{articleService: articleService}
}

// CreateArticle 创建文章
func (c *ArticleController) CreateArticle(ctx *gin.Context) {
	var input struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Preview    string `json:"preview" binding:"required"`
		CategoryID *uint  `json:"category_id"` // 新增分类 ID
		TagIDs     []uint `json:"tag_ids"`     // 新增标签 ID 数组
		CoverImg   string `json:"cover_img"`   //【新增】封面图的 URL
	}

	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "用户未登录或身份已过期"})
		return
	}
	bodyBytes, _ := io.ReadAll(ctx.Request.Body)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := ctx.ShouldBindJSON(&input); err != nil {
		log.Printf("绑定失败，原始 JSON: %s", string(bodyBytes))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article := models.Article{
		Title:      input.Title,
		Content:    input.Content,
		Preview:    input.Preview,
		CategoryID: input.CategoryID,
		CoverImg:   input.CoverImg,
		UserID:     id.(uint),
	}

	if err := c.articleService.CreateArticle(ctx.Request.Context(), &article, input.TagIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func(title string, articleId uint, content string) {
		grpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		artIDStr := strconv.FormatUint(uint64(articleId), 10)
		rpcResp, err := global.Rag_grpc_client.AddRag(grpcCtx, &rag_grpc.RagRequest{
			Title:     title,
			ArticleId: artIDStr,
			Content:   content,
		})
		if err != nil {
			// 异步任务失败不能影响用户，打印错误日志供后续排查/重试队列处理即可
			log.Printf("[RAG同步失败] 文章ID: %s, 错误原因: %v", artIDStr, err)
		}
		if !rpcResp.Ok {
			log.Printf("[RAG同步业务失败] 文章ID: %s, Python端处理失败", artIDStr)
		}
		fmt.Printf("[RAG同步成功] 文章ID: %s 已成功在向量库建索", artIDStr)
	}(article.Title, article.ID, article.Content)

	ctx.JSON(http.StatusCreated, article)
}

// GetArticles 获取所有文章
func (c *ArticleController) GetArticles(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	categoryIDStr := ctx.Query("category_id") // 新增：获取分类筛选参数
	tagIDStr := ctx.Query("tag_id")           // 新增：获取标签筛选参数
	keyword := ctx.Query("keyword")           // 新增：获取搜索关键词参数
	userIDStr := ctx.Query("user_id")         // 新增：按用户ID筛选
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	categoryID, _ := strconv.Atoi(categoryIDStr)
	tagID, _ := strconv.Atoi(tagIDStr)
	userID, _ := strconv.Atoi(userIDStr)

	cacheObj, page, pageSize, total, err := c.articleService.GetArticles(ctx.Request.Context(), page, pageSize, categoryID, tagID, keyword, uint(userID))
	if err != nil {
		if err.Error() == "请求页码超出最大支持范围" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize) //总页数，(21+5-1)/5=5
	//返回给前端
	ctx.JSON(http.StatusOK, gin.H{
		"data":        cacheObj.Data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetArticlesByID 根据ID获取文章
func (c *ArticleController) GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	article, comments, likes, err := c.articleService.GetArticlesByID(ctx.Request.Context(), id)
	if err != nil {
		if err.Error() == "该文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "数据组装失败", "detail": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"article":  article,
		"comments": comments,
		"likes":    likes,
	})
}

// DelArticle 删除文章
func (c *ArticleController) DelArticle(ctx *gin.Context) {
	idA := ctx.Param("id")
	idU, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	if err := c.articleService.DelArticle(ctx.Request.Context(), idA, idU.(uint)); err != nil {
		if err.Error() == "文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权限删除该文章" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "message": "删除成功"})
}

// UpdateArticle 更新文章
func (c *ArticleController) UpdateArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")
	userID, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

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

	article, err := c.articleService.UpdateArticle(ctx.Request.Context(), articleID, userID.(uint), struct {
		Title      string
		Content    string
		Preview    string
		CategoryID uint
		TagIDs     []uint
		CoverImg   string
	}{
		Title:      input.Title,
		Content:    input.Content,
		Preview:    input.Preview,
		CategoryID: input.CategoryID,
		TagIDs:     input.TagIDs,
		CoverImg:   input.CoverImg,
	})
	if err != nil {
		if err.Error() == "文章不存在" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "无权修改他人的文章" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "更新成功", "article": article})
}

// GetArticlesByCursor 获取ID小于cursor的limit个Article
// 请求：GET /articles?cursor=12345&limit=10
// 返回：最后一条的 ID 作为下次请求的 cursor
func (c *ArticleController) GetArticlesByCursor(ctx *gin.Context) {
	cursorStr := ctx.DefaultQuery("cursor", "0")
	limitStr := ctx.DefaultQuery("limit", "10")

	cursor, _ := strconv.ParseUint(cursorStr, 10, 64)
	limit, _ := strconv.Atoi(limitStr)

	articles, nextCursor, hasMore, err := c.articleService.GetArticlesByCursor(ctx.Request.Context(), cursor, limit)
	if err != nil {
		log.Println("获取文章列表失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":        articles,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

func (c *ArticleController) SearchRagArticle(ctx *gin.Context) {
	type RagItemVO struct {
		ArticleID string `json:"article_id"`
		Title     string `json:"title"`
	}

	keyword := ctx.Query("keyword")
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}
	response, err := global.Rag_grpc_client.SearchRag(ctx.Request.Context(), &rag_grpc.RagSearchRequest{Query: keyword})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	articleList := make([]RagItemVO, 0, len(response.GetItems()))
	for _, item := range response.GetItems() {
		articleList = append(articleList, RagItemVO{
			ArticleID: item.GetArticleId(), // 使用 Getter 防空指针
			Title:     item.GetTitle(),
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": articleList,
	})
}
