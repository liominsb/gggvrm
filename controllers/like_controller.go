package controllers

import (
	"errors"
	"gggvrm/global"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// LikeArticle 赞成文章
func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	if err := global.RedisDB.Incr(likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法增加点赞数"})
		return
	}

	if err:=global.Db.

	ctx.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

func GetArticlelikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	likes, err := global.RedisDB.Get(likeKey).Result()

	if errors.Is(err, redis.Nil) {
		likes = "0"
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取点赞数"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
