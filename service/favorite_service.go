package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/models"
	"gggvrm/repository"
	"gggvrm/utils"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// FavoriteService 收藏业务接口
type FavoriteService interface {
	ToggleFavorite(ctx context.Context, articleIDStr string, userID uint) (bool, error)                     // 切换收藏状态（已收藏则取消，未收藏则添加）
	IsFavorited(ctx context.Context, articleIDStr string, userID uint) (bool, error)                        // 查询当前用户是否已收藏该文章
	GetFavoriteCount(ctx context.Context, articleIDStr string) (int64, error)                               // 获取文章的收藏总数
	GetUserFavorites(ctx context.Context, userID uint, page, pageSize int) ([]models.Article, int64, error) // 分页获取用户的收藏文章列表
}

type favoriteServiceImpl struct {
	favoriteRepo repository.FavoriteRepository
	redisClient  *redis.Client
}

// NewFavoriteRepository 创建收藏服务实例，注入收藏仓库和 Redis 客户端
func NewFavoriteService(favoriteRepo repository.FavoriteRepository, redisClient *redis.Client) FavoriteService {
	return &favoriteServiceImpl{
		favoriteRepo: favoriteRepo,
		redisClient:  redisClient,
	}
}

// ToggleFavorite 切换收藏状态：查询数据库判断当前状态，已收藏则删除记录并 Redis -1，未收藏则插入记录并 Redis +1，收藏时通过 MQ 异步通知
func (s *favoriteServiceImpl) ToggleFavorite(ctx context.Context, articleIDStr string, userID uint) (bool, error) {
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		return false, errors.New("无效的文章ID")
	}

	favKey := fmt.Sprintf("article:%d:favorites", articleID)
	isFavKey := fmt.Sprintf("user:%d:fav:%d", userID, articleID)

	// 查询当前收藏状态
	isFav, err := s.favoriteRepo.IsFavorited(ctx, userID, uint(articleID))
	if err != nil {
		return false, errors.New("查询收藏状态失败")
	}

	if isFav {
		// 已收藏，执行取消收藏
		if err := s.favoriteRepo.RemoveFavorite(ctx, userID, uint(articleID)); err != nil {
			return false, errors.New("取消收藏失败")
		}

		// Redis 收藏数 -1
		newCount, err := s.redisClient.Decr(ctx, favKey).Result()
		if err != nil {
			log.Printf("【警告】Redis 收藏数递减失败 Key: %s, Err: %v", favKey, err)
		}
		if newCount < 0 {
			s.redisClient.Set(ctx, favKey, 0, utils.RandomExpiration(10*time.Minute))
		}
		s.redisClient.Del(ctx, isFavKey)
		s.clearUserFavoriteListCache(ctx, userID)

		// 清除用户收藏状态缓存
		s.redisClient.Del(ctx, isFavKey)

		// 通过 MQ 异步通知（可选，取消收藏不一定要通知）
		return false, nil
	}

	// 未收藏，执行添加收藏
	if err := s.favoriteRepo.AddFavorite(ctx, userID, uint(articleID)); err != nil {
		return false, errors.New("收藏失败")
	}

	// 确保 Redis 中有收藏数缓存
	exists, err := s.redisClient.Exists(ctx, favKey).Result()
	if err != nil {
		log.Printf("【警告】Redis 检查收藏数缓存异常: %v", err)
	}
	if exists == 0 {
		count, err := s.favoriteRepo.GetFavoriteCount(ctx, uint(articleID))
		if err != nil {
			log.Printf("【警告】获取收藏数失败: %v", err)
		}
		if err := s.redisClient.SetNX(ctx, favKey, count, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
			log.Printf("【警告】初始化收藏数缓存失败: %v", err)
		}
	}

	// Redis 收藏数 +1
	_, err = s.redisClient.Incr(ctx, favKey).Result()
	s.redisClient.Set(ctx, isFavKey, 1, utils.RandomExpiration(10*time.Minute))
	s.clearUserFavoriteListCache(ctx, userID)
	if err != nil {
		log.Printf("【警告】Redis 收藏数递增失败 Key: %s, Err: %v", favKey, err)
	}

	// 设置用户收藏状态缓存
	s.redisClient.Set(ctx, isFavKey, 1, utils.RandomExpiration(10*time.Minute))

	//// 消息体富化：生产者查好用户名，消费者零查库
	//authorUsername, _ := s.favoriteRepo.GetArticleAuthorUsername(ctx, uint(articleID))
	//
	//msgData, _ := json.Marshal(map[string]interface{}{
	//	"action":          "favorite_article",
	//	"article_id":      articleID,
	//	"user_id":         userID,
	//	"author_username": authorUsername,
	//	"timestamp":       time.Now().Unix(),
	//})
	//
	//err = mq.PublishMessage("favorite_tasks", msgData)
	//if err != nil {
	//	log.Printf("【RabbitMQ警告】发送收藏消息失败: %v\n", err)
	//}

	return true, nil
}

// IsFavorited 查询收藏状态：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *favoriteServiceImpl) clearUserFavoriteListCache(ctx context.Context, userID uint) {
	var cursor uint64
	pattern := fmt.Sprintf("user:%d:favorites:page:*", userID)
	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("clear user favorite cache failed, userID=%d: %v", userID, err)
			return
		}
		if len(keys) > 0 {
			s.redisClient.Del(ctx, keys...)
		}
		if nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func (s *favoriteServiceImpl) IsFavorited(ctx context.Context, articleIDStr string, userID uint) (bool, error) {
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		return false, errors.New("无效的文章ID")
	}

	isFavKey := fmt.Sprintf("user:%d:fav:%d", userID, articleID)

	// 先查 Redis 缓存
	val, err := s.redisClient.Get(ctx, isFavKey).Result()
	if err == nil {
		return val == "1", nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询收藏状态失败:", err)
	}

	// 缓存未命中，查数据库
	isFav, err := s.favoriteRepo.IsFavorited(ctx, userID, uint(articleID))
	if err != nil {
		return false, errors.New("查询收藏状态失败")
	}

	// 回写缓存
	cacheVal := "0"
	if isFav {
		cacheVal = "1"
	}
	if err := s.redisClient.Set(ctx, isFavKey, cacheVal, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
		log.Printf("【警告】回写收藏状态缓存失败 Key: %s, Err: %v", isFavKey, err)
	}

	return isFav, nil
}

// GetFavoriteCount 获取收藏数：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *favoriteServiceImpl) GetFavoriteCount(ctx context.Context, articleIDStr string) (int64, error) {
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil || articleID <= 0 {
		return 0, errors.New("无效的文章ID")
	}

	favKey := fmt.Sprintf("article:%d:favorites", articleID)

	// 先查 Redis 缓存
	countStr, err := s.redisClient.Get(ctx, favKey).Result()
	if err == nil {
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err == nil {
			return count, nil
		}
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询收藏数失败:", err)
	}

	// 缓存未命中，查数据库
	count, err := s.favoriteRepo.GetFavoriteCount(ctx, uint(articleID))
	if err != nil {
		return 0, errors.New("查询收藏数失败")
	}

	// 回写缓存
	if err := s.redisClient.Set(ctx, favKey, count, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
		log.Printf("【警告】回写收藏数缓存失败 Key: %s, Err: %v", favKey, err)
	}

	return count, nil
}

// GetUserFavorites 获取用户收藏列表：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *favoriteServiceImpl) GetUserFavorites(ctx context.Context, userID uint, page, pageSize int) ([]models.Article, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("user:%d:favorites:page:%d:size:%d", userID, page, pageSize)

	// 尝试从缓存获取
	cacheData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var result struct {
			Articles []models.Article `json:"articles"`
			Total    int64            `json:"total"`
		}
		if json.Unmarshal([]byte(cacheData), &result) == nil {
			return result.Articles, result.Total, nil
		}
		log.Println("收藏列表缓存解析失败，降级到数据库查询")
	} else if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询收藏列表失败:", err)
	}

	// 缓存未命中，查数据库
	articles, total, err := s.favoriteRepo.GetUserFavorites(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, errors.New("查询收藏列表失败")
	}

	// 回写缓存
	result := struct {
		Articles []models.Article `json:"articles"`
		Total    int64            `json:"total"`
	}{
		Articles: articles,
		Total:    total,
	}
	if err := utils.Setcache(ctx, cacheKey, result); err != nil {
		log.Printf("【警告】回写收藏列表缓存失败 Key: %s, Err: %v", cacheKey, err)
	}

	return articles, total, nil
}
