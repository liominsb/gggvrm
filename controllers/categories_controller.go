package controllers

import (
	"gggvrm/models"
	"gggvrm/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CateController struct {
	cateService service.CateService
}

func NewCateController(cateService service.CateService) *CateController {
	return &CateController{cateService: cateService}
}

func (c *CateController) CreateCate(ctx *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"参数错误": err.Error()})
		return
	}

	cate := &models.Category{
		Name: input.Name,
	}
	if err := c.cateService.CreateCate(ctx.Request.Context(), cate); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"创建Category失败,error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, cate)
}

func (c *CateController) GetCates(ctx *gin.Context) {
	var cates []models.Category
	result, err := c.cateService.GetCates(ctx.Request.Context(), &cates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"获取Categories失败,error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *CateController) DeleteCate(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"删除Category失败,error": err.Error()})
		return
	}
	if err := c.cateService.DeleteCate(ctx.Request.Context(), uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"删除Category失败,error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
