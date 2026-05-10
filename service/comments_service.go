package service

import (
	"context"
	"encoding/json"
	"errors"
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
	cacheData, err := s.redisClient.Get(ctx, cacheKey).Result()
	var comments []models.Comment

	// 1. 处理 Redis 读取结果
	if err == nil {
		if jsonErr := json.Unmarshal([]byte(cacheData), &comments); jsonErr == nil {
			return comments, nil // 正常命中缓存
		}
		log.Println("缓存数据解析失败，尝试降级查库:", err)
	} else if !errors.Is(err, redis.Nil) {
		// 如果不是缓存未命中，而是 Redis 真的出错了（比如网络异常），记录错误日志
		log.Printf("Redis 读取异常 (key: %s): %v\n", cacheKey, err)
		// 视你的业务容忍度，这里可以选择直接 return nil, err 保护 DB，或者继续向下走降级查库
	}

	// 2. 缓存未命中或解析失败，使用 Singleflight 进行查库拦截
	v, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {

		// 3. 执行真正的数据库查询
		dbComments, dbErr := s.commentsRepo.GetComments(ctx, articleID)
		if dbErr != nil {
			// 修正1：查库失败必须把错误返回，绝对不能写缓存，也不能 return nil
			return nil, dbErr
		}

		// 4. 查库成功后，回写缓存
		if setErr := utils.Setcache(ctx, cacheKey, dbComments); setErr != nil {
			log.Printf("【警告】设置评论缓存失败 cacheKey: %s, err: %v\n", cacheKey, setErr)
		}

		return dbComments, nil
	})

	// 5. 统一处理 Singleflight 闭包中返回的错误
	if err != nil {
		return nil, err // 把 DB 的真实错误抛给 Controller 层
	}

	return v.([]models.Comment), nil
}
