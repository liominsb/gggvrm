package controllers

import (
	"errors"
	"gggvrm/global"
	"gggvrm/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// LikeArticle 赞成文章
func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	exists, err := global.RedisDB.Exists(likeKey).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法检查点赞数"})
		return
	}
	if exists == 0 {
		var article models.Article
		// 只查询 likes 字段，提高效率
		if err := global.Db.Select("likes").Where("id = ?", articleID).First(&article).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}

		if err := global.RedisDB.SetNX(likeKey, article.Likes, 0).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法初始化点赞数"})
			return
		}
	}

	if err := global.RedisDB.Incr(likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法增加点赞数"})
		return
	}

	//msgData, _ := json.Marshal(map[string]interface{}{
	//	"action":     "like_article",
	//	"article_id": articleID,
	//})
	//
	//err = global.RabbitMQChan.Publish(
	//	"",           // 默认交换机
	//	"like_tasks", // 你的队列名称
	//	false,
	//	false,
	//	amqp.Publishing{
	//		ContentType: "application/json",
	//		Body:        msgData,
	//	},
	//)
	//if err != nil {
	//	fmt.Printf("【RabbitMQ警告】发送点赞消息失败: %v\n", err)
	//}

	ctx.JSON(http.StatusOK, gin.H{"message": "点赞成功"})
}

func GetArticlelikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":likes"

	likes, err := global.RedisDB.Get(likeKey).Result()

	if errors.Is(err, redis.Nil) {
		var article models.Article
		// 只查询 likes 字段，提高效率
		if err := global.Db.Select("likes").Where("id = ?", articleID).First(&article).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
			return
		}

		if err := global.RedisDB.SetNX(likeKey, article.Likes, 0).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法初始化点赞数"})
			return
		}
		likes = strconv.Itoa(article.Likes)
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取点赞数"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"likes": likes})
}
