package controllers

import (
	"fmt"
	"gggvrm/models"
	"gggvrm/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	commentService service.CommentService
}

// 构造函数：在外部（比如 main.go 的路由配置里）初始化时传入依赖
func NewcCommentController(commentService service.CommentService) *CommentController {
	return &CommentController{commentService: commentService}
}

// CreateComments 创建评论
func (c *CommentController) CreateComment(ctx *gin.Context) {
	articleIDstring := ctx.Param("id")

	articleID, err := strconv.ParseUint(articleIDstring, 10, 32)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	cacheKey := fmt.Sprintf("article:%d:comments", articleID)

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

	err = c.commentService.CreateComment(ctx, cacheKey, &comment)
	if err != nil {
		return
	}

	//msgData, _ := json.Marshal(map[string]interface{}{
	//	"action":     "new_comment",
	//	"article_id": articleID,
	//	"user_id":    userid,
	//})
	//mq.PublishMessage("article_tasks", msgData)

	ctx.JSON(http.StatusCreated, comment)
}

func (c *CommentController) DelComment(ctx *gin.Context) {
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

	err = c.commentService.DelComment(ctx, commentID, userid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetComments 获取单个文章的所有评论
func (c *CommentController) GetComments(ctx *gin.Context) {

	articleIDstring := ctx.Param("id")
	articleID, err := strconv.ParseUint(articleIDstring, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	comments, err := c.commentService.GetComments(ctx, uint(articleID))
	if err != nil {
		log.Println("获取评论失败:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}

	ctx.JSON(http.StatusOK, comments)
}
