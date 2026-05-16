package service

import (
	"context"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/repository"
	"gggvrm/utils"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type CommentService interface {
	CreateComment(ctx context.Context, cacheKey string, comment *models.Comment) error
	DelComment(ctx context.Context, commentID uint64, userid uint) error
	GetComments(ctx context.Context, articleID uint) ([]models.Comment, error)
}

type commentServiceImpl struct {
	commentsRepo repository.CommentsRepository
	redisClient  *redis.Client
	sfGroup      singleflight.Group
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

func (s *commentServiceImpl) DelComment(ctx context.Context, commentID uint64, userid uint) error {
	comment, err := s.commentsRepo.GetCommentByID(ctx, uint(commentID))
	if err != nil {
		return err
	}
	article, err := s.commentsRepo.GetArticleByID(ctx, comment.ArticleID)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("article:%d:comments", article.ID)

	if userid != comment.UserID && userid != article.UserID {
		log.Println("【错误】用户没有权限删除评论")
		return fmt.Errorf("用户没有权限删除评论")
	}

	//第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		fmt.Printf("【Redis警告】清理文章 %d 缓存失败: %v\n", comment.ArticleID, err)
	}

	err = s.commentsRepo.DelComment(ctx, comment)
	if err != nil {
		log.Println("数据库删除评论失败:", err)
		return fmt.Errorf("删除评论失败: %w", err)
	}

	go func(bgctx context.Context) {
		time.Sleep(100 * time.Millisecond) //延时
		//第二次删除缓存
		if err := s.redisClient.Del(bgctx, cacheKey).Err(); err != nil {
			fmt.Printf("【Redis警告】清理文章 %d 缓存失败: %v\n", comment.ArticleID, err)
		}
	}(context.Background())
	return nil
}

func (s *commentServiceImpl) GetComments(ctx context.Context, articleID uint) ([]models.Comment, error) {
	cacheKey := fmt.Sprintf("article:%d:comments", articleID)

	// 使用 Singleflight 防止缓存击穿，内部使用 GetCacheOrQuery 处理缓存逻辑
	v, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*[]models.Comment, error) {
			comments, err := s.commentsRepo.GetComments(ctx, articleID)
			if err != nil {
				return nil, err
			}
			return &comments, nil
		})
		if err != nil {
			return nil, err
		}
		return *result, nil
	})

	if err != nil {
		return nil, err
	}

	return v.([]models.Comment), nil
}
