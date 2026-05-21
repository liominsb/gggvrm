package controllers

import (
	"gggvrm/models"
	"gggvrm/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 定义一个 Controller 结构体，用于持有依赖
type TagsController struct {
	tagsService service.TagsService
}

// 构造函数：在外部（比如 main.go 的路由配置里）初始化时传入依赖
func NewTagsController(tagsService service.TagsService) *TagsController {
	return &TagsController{tagsService: tagsService}
}

func (c *TagsController) CreateTag(ctx *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tag := &models.Tag{
		Name: input.Name,
	}
	if err := c.tagsService.CreateTag(ctx.Request.Context(), tag); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"创建Tag失败,error": err.Error()})
	}
	ctx.JSON(http.StatusOK, tag)
}

func (c *TagsController) GetTags(ctx *gin.Context) {
	var tags []models.Tag
	if err := c.tagsService.GetTags(ctx, &tags); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"获取Tags失败,error": err.Error()})
	}
	ctx.JSON(http.StatusOK, tags)
}

func (c *TagsController) DeleteTag(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"删除Tag失败,error": err.Error()})
	}
	if err := c.tagsService.DelTag(ctx.Request.Context(), uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"删除Tag失败,error": err.Error()})
	}
	ctx.JSON(http.StatusOK, nil)
}
