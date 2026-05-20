package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/models"
	"gggvrm/mq"
	"gggvrm/repository"
	"gggvrm/utils"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type LikeService interface {
	GetArticleLikes(ctx context.Context, articleIDStr string) (string, error)
	LikeArticle(ctx context.Context, articleIDStr string, userID uint) (string, error)
}

type likeServiceImpl struct {
	likeRepo    repository.LikeRepository
	redisClient *redis.Client
}

func NewLikeService(likeRepo repository.LikeRepository, redisClient *redis.Client) LikeService {
	return &likeServiceImpl{
		likeRepo:    likeRepo,
		redisClient: redisClient,
	}
}

func (s *likeServiceImpl) GetArticleLikes(ctx context.Context, articleIDStr string) (string, error) {
	// 1. 业务校验：验证传入的 ID 是否合法
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		return "", errors.New("无效的文章ID")
	}

	likeKey := fmt.Sprintf("article:%d:likes", articleID)

	// 1. 尝试从 Redis 获取
	likesStr, err := s.redisClient.Get(ctx, likeKey).Result()
	if err == nil {
		return likesStr, nil
	}

	if !errors.Is(err, redis.Nil) {
		// Redis 发生其他异常
		log.Println("Redis 获取点赞数失败:", err)
	}

	// 2. 调用底层获取数据
	likes, err := s.likeRepo.GetArticleLikes(ctx, uint(articleID))
	if err != nil {
		return "", err
	}

	if err := s.redisClient.Set(ctx, likeKey, likes, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
		log.Printf("【警告】点赞数回写Redis失败 Key: %s, Err: %v", likeKey, err)
	}
	return strconv.Itoa(likes), nil
}

func (s *likeServiceImpl) LikeArticle(ctx context.Context, articleIDStr string, userID uint) (string, error) {
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		return "", errors.New("无效的文章ID")
	}

	likeKey := fmt.Sprintf("article:%d:likes", articleID)

	exists, err := s.redisClient.Exists(ctx, likeKey).Result()
	if err != nil {
		return "0", errors.New("检查缓存异常")
	}
	if exists == 0 {

		likes, err := s.likeRepo.GetArticleLikes(ctx, uint(articleID))
		if err != nil {
			return "", err
		}

		if err := s.redisClient.SetNX(ctx, likeKey, likes, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
			return "0", errors.New("无法初始化点赞数缓存")
		}
	}

	article := models.Article{}
	cacheKey := fmt.Sprintf("article:detail:%d", articleID)
	result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.Article, error) {
		if err := s.likeRepo.GetArticleByIDWithPreload(ctx, &article, strconv.Itoa(articleID)); err != nil {
			return nil, err
		}
		return &article, nil
	})
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", err
	}
	article = *result

	newLikes, err := s.redisClient.Incr(ctx, likeKey).Result()
	if err != nil {
		return "0", errors.New("点赞失败")
	}

	msgData, _ := json.Marshal(map[string]interface{}{
		"action":         "like_article",
		"article_id":     articleID,
		"user_id":        userID, // 新增：谁点的赞（核心参数）
		"AuthorUsername": article.User.Username,
		"timestamp":      time.Now().Unix(), // 新增：事件发生时间
	})

	// 使用你之前在 mq/producer.go 中封装好的函数
	err = mq.PublishMessage("like_tasks", msgData)
	if err != nil {
		// 发送失败只需打日志，千万别 return err，不能因为发消息失败导致用户的点赞操作失败（保证核心主流程可用）
		log.Printf("【RabbitMQ警告】发送点赞消息失败: %v\n", err)
	}

	return strconv.Itoa(int(newLikes)), nil
}
