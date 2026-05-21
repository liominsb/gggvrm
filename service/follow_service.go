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
	"time"

	"github.com/redis/go-redis/v9"
)

// FollowService 关注业务接口
type FollowService interface {
	ToggleFollow(ctx context.Context, followeeIDStr string, followerID uint) (bool, error)           // 切换关注状态（已关注则取消，未关注则添加）
	IsFollowing(ctx context.Context, followeeIDStr string, followerID uint) (bool, error)            // 查询当前用户是否已关注目标用户
	GetFollowCounts(ctx context.Context, userIDStr string) (int64, int64, error)                     // 获取用户的关注数和粉丝数
	GetFollowing(ctx context.Context, userID uint, page, pageSize int) ([]models.User, int64, error) // 分页获取用户的关注列表
	GetFollowers(ctx context.Context, userID uint, page, pageSize int) ([]models.User, int64, error) // 分页获取用户的粉丝列表
}

type followServiceImpl struct {
	followRepo  repository.FollowRepository
	redisClient *redis.Client
}

// NewFollowService 创建关注服务实例，注入关注仓库和 Redis 客户端
func NewFollowService(followRepo repository.FollowRepository, redisClient *redis.Client) FollowService {
	return &followServiceImpl{
		followRepo:  followRepo,
		redisClient: redisClient,
	}
}

// ToggleFollow 切换关注状态：查询数据库判断当前状态，已关注则删除记录并双方计数 -1，未关注则插入记录并双方计数 +1，关注时通过 MQ 异步通知
func (s *followServiceImpl) ToggleFollow(ctx context.Context, followeeIDStr string, followerID uint) (bool, error) {
	var followeeID uint
	if _, err := fmt.Sscanf(followeeIDStr, "%d", &followeeID); err != nil || followeeID == 0 {
		return false, errors.New("无效的用户ID")
	}

	// 不能关注自己
	if followerID == followeeID {
		return false, errors.New("不能关注自己")
	}

	followingKey := fmt.Sprintf("user:%d:following_count", followerID)
	followersKey := fmt.Sprintf("user:%d:followers_count", followeeID)
	isFollowKey := fmt.Sprintf("user:%d:follow:%d", followerID, followeeID)

	// 查询当前关注状态
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return false, errors.New("查询关注状态失败")
	}

	if isFollowing {
		// 已关注，执行取消关注
		if err := s.followRepo.Unfollow(ctx, followerID, followeeID); err != nil {
			return false, errors.New("取消关注失败")
		}

		// Redis 关注数 -1
		newFollowingCount, err := s.redisClient.Decr(ctx, followingKey).Result()
		if err != nil {
			log.Printf("【警告】Redis 关注数递减失败 Key: %s, Err: %v", followingKey, err)
		}
		if newFollowingCount < 0 {
			s.redisClient.Set(ctx, followingKey, 0, utils.RandomExpiration(10*time.Minute))
		}

		// Redis 粉丝数 -1
		newFollowersCount, err := s.redisClient.Decr(ctx, followersKey).Result()
		if err != nil {
			log.Printf("【警告】Redis 粉丝数递减失败 Key: %s, Err: %v", followersKey, err)
		}
		if newFollowersCount < 0 {
			s.redisClient.Set(ctx, followersKey, 0, utils.RandomExpiration(10*time.Minute))
		}

		// 清除关注状态缓存
		s.redisClient.Del(ctx, isFollowKey)

		return false, nil
	}

	// 未关注，执行关注
	if err := s.followRepo.Follow(ctx, followerID, followeeID); err != nil {
		return false, errors.New("关注失败")
	}

	// 确保 Redis 中有关注数缓存
	s.ensureCountCache(ctx, followingKey, func() (int64, error) {
		return s.followRepo.GetFollowingCount(ctx, followerID)
	})
	s.ensureCountCache(ctx, followersKey, func() (int64, error) {
		return s.followRepo.GetFollowersCount(ctx, followeeID)
	})

	// Redis 关注数 +1
	_, err = s.redisClient.Incr(ctx, followingKey).Result()
	if err != nil {
		log.Printf("【警告】Redis 关注数递增失败 Key: %s, Err: %v", followingKey, err)
	}

	// Redis 粉丝数 +1
	_, err = s.redisClient.Incr(ctx, followersKey).Result()
	if err != nil {
		log.Printf("【警告】Redis 粉丝数递增失败 Key: %s, Err: %v", followersKey, err)
	}

	// 设置关注状态缓存
	s.redisClient.Set(ctx, isFollowKey, 1, utils.RandomExpiration(10*time.Minute))

	//// 消息体富化：生产者查好用户名，消费者零查库
	//followerUsername, _ := s.followRepo.GetUsername(ctx, followerID)
	//
	//msgData, _ := json.Marshal(map[string]interface{}{
	//	"action":            "follow_user",
	//	"follower_id":       followerID,
	//	"follower_username": followerUsername,
	//	"followee_id":       followeeID,
	//	"timestamp":         time.Now().Unix(),
	//})
	//
	//err = mq.PublishMessage("follow_tasks", msgData)
	//if err != nil {
	//	log.Printf("【RabbitMQ警告】发送关注消息失败: %v\n", err)
	//}

	return true, nil
}

// IsFollowing 查询关注状态：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *followServiceImpl) IsFollowing(ctx context.Context, followeeIDStr string, followerID uint) (bool, error) {
	var followeeID uint
	if _, err := fmt.Sscanf(followeeIDStr, "%d", &followeeID); err != nil || followeeID == 0 {
		return false, errors.New("无效的用户ID")
	}

	isFollowKey := fmt.Sprintf("user:%d:follow:%d", followerID, followeeID)

	// 先查 Redis 缓存
	val, err := s.redisClient.Get(ctx, isFollowKey).Result()
	if err == nil {
		return val == "1", nil
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询关注状态失败:", err)
	}

	// 缓存未命中，查数据库
	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followeeID)
	if err != nil {
		return false, errors.New("查询关注状态失败")
	}

	// 回写缓存
	cacheVal := "0"
	if isFollowing {
		cacheVal = "1"
	}
	if err := s.redisClient.Set(ctx, isFollowKey, cacheVal, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
		log.Printf("【警告】回写关注状态缓存失败 Key: %s, Err: %v", isFollowKey, err)
	}

	return isFollowing, nil
}

// GetFollowCounts 获取关注数和粉丝数：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *followServiceImpl) GetFollowCounts(ctx context.Context, userIDStr string) (int64, int64, error) {
	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil || userID == 0 {
		return 0, 0, errors.New("无效的用户ID")
	}

	followingKey := fmt.Sprintf("user:%d:following_count", userID)
	followersKey := fmt.Sprintf("user:%d:followers_count", userID)

	// 先查 Redis 缓存
	followingCount := s.getCountFromCache(ctx, followingKey)
	followersCount := s.getCountFromCache(ctx, followersKey)

	if followingCount >= 0 && followersCount >= 0 {
		return followingCount, followersCount, nil
	}

	// 缓存未命中，查数据库
	dbFollowingCount, err := s.followRepo.GetFollowingCount(ctx, userID)
	if err != nil {
		return 0, 0, errors.New("查询关注数失败")
	}

	dbFollowersCount, err := s.followRepo.GetFollowersCount(ctx, userID)
	if err != nil {
		return 0, 0, errors.New("查询粉丝数失败")
	}

	// 回写缓存
	s.redisClient.Set(ctx, followingKey, dbFollowingCount, utils.RandomExpiration(10*time.Minute))
	s.redisClient.Set(ctx, followersKey, dbFollowersCount, utils.RandomExpiration(10*time.Minute))

	return dbFollowingCount, dbFollowersCount, nil
}

// GetFollowing 获取关注列表：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *followServiceImpl) GetFollowing(ctx context.Context, userID uint, page, pageSize int) ([]models.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("user:%d:following:page:%d:size:%d", userID, page, pageSize)

	// 尝试从缓存获取
	cacheData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var result struct {
			Users []models.User `json:"users"`
			Total int64         `json:"total"`
		}
		if json.Unmarshal([]byte(cacheData), &result) == nil {
			return result.Users, result.Total, nil
		}
		log.Println("关注列表缓存解析失败，降级到数据库查询")
	} else if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询关注列表失败:", err)
	}

	// 缓存未命中，查数据库
	users, total, err := s.followRepo.GetFollowing(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, errors.New("查询关注列表失败")
	}

	// 回写缓存
	result := struct {
		Users []models.User `json:"users"`
		Total int64         `json:"total"`
	}{
		Users: users,
		Total: total,
	}
	if err := utils.Setcache(ctx, cacheKey, result); err != nil {
		log.Printf("【警告】回写关注列表缓存失败 Key: %s, Err: %v", cacheKey, err)
	}

	return users, total, nil
}

// GetFollowers 获取粉丝列表：先查 Redis 缓存，缓存未命中则查数据库并回写缓存
func (s *followServiceImpl) GetFollowers(ctx context.Context, userID uint, page, pageSize int) ([]models.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	cacheKey := fmt.Sprintf("user:%d:followers:page:%d:size:%d", userID, page, pageSize)

	// 尝试从缓存获取
	cacheData, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var result struct {
			Users []models.User `json:"users"`
			Total int64         `json:"total"`
		}
		if json.Unmarshal([]byte(cacheData), &result) == nil {
			return result.Users, result.Total, nil
		}
		log.Println("粉丝列表缓存解析失败，降级到数据库查询")
	} else if !errors.Is(err, redis.Nil) {
		log.Println("Redis 查询粉丝列表失败:", err)
	}

	// 缓存未命中，查数据库
	users, total, err := s.followRepo.GetFollowers(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, errors.New("查询粉丝列表失败")
	}

	// 回写缓存
	result := struct {
		Users []models.User `json:"users"`
		Total int64         `json:"total"`
	}{
		Users: users,
		Total: total,
	}
	if err := utils.Setcache(ctx, cacheKey, result); err != nil {
		log.Printf("【警告】回写粉丝列表缓存失败 Key: %s, Err: %v", cacheKey, err)
	}

	return users, total, nil
}

// ensureCountCache 确保 Redis 中有计数缓存
func (s *followServiceImpl) ensureCountCache(ctx context.Context, key string, queryFunc func() (int64, error)) {
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("【警告】Redis 检查缓存异常 Key: %s, Err: %v", key, err)
		return
	}
	if exists == 0 {
		count, err := queryFunc()
		if err != nil {
			log.Printf("【警告】查询计数失败 Key: %s, Err: %v", key, err)
			return
		}
		if err := s.redisClient.SetNX(ctx, key, count, utils.RandomExpiration(10*time.Minute)).Err(); err != nil {
			log.Printf("【警告】初始化计数缓存失败 Key: %s, Err: %v", key, err)
		}
	}
}

// getCountFromCache 从 Redis 获取计数，返回 -1 表示缓存未命中
func (s *followServiceImpl) getCountFromCache(ctx context.Context, key string) int64 {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return -1
	}
	var count int64
	if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
		return -1
	}
	return count
}
