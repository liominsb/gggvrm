package service

import (
	"context"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/repository"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type CommentService interface {
	CreateComment(ctx context.Context, cacheKey string, comment *models.Comment) error
}

type commentServiceImpl struct {
	commentsRepo repository.CommentsRepository
	redisClient  *redis.Client
}

func NewCommentService(commentsRepo repository.CommentsRepository, redisClient *redis.Client) CommentService {
	return &commentServiceImpl{commentsRepo: commentsRepo, redisClient: redisClient}
}

func (s *commentServiceImpl) CreateComment(ctx context.Context, cacheKey string, comment *models.Comment) error {
	//第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("【警告】第一次删除缓存失败: %s, err: %v\n", cacheKey, err)
	}

	err := s.commentsRepo.CreateComment(ctx, comment)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(100 * time.Millisecond) //延时

		//第二次删除缓存
		if err := global.RedisDB.Del(context.Background(), cacheKey).Err(); err != nil {
			log.Printf("【警告】延时双删失败 cacheKey: %s, err: %v\n", cacheKey, err)
			return
		}
	}()
	return nil
}
