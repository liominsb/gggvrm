package service

import (
	"context"
	"fmt"
	"gggvrm/models"
	"gggvrm/repository"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// FeedService 个人 Feed 流服务接口
type FeedService interface {
	GetUserFeed(ctx context.Context, userID uint, page, pageSize int) ([]ArticleListResponse, int64, error)
}

type feedServiceImpl struct {
	articleRepo repository.ArticleRepository
	redisClient *redis.Client
}

func NewFeedService(articleRepo repository.ArticleRepository, redisClient *redis.Client) FeedService {
	return &feedServiceImpl{articleRepo: articleRepo, redisClient: redisClient}
}

// GetUserFeed 从 Redis ZSet 中取出当前登录用户 Feed 流的文章 ID，再查 MySQL 获取具体内容
func (s *feedServiceImpl) GetUserFeed(ctx context.Context, userID uint, page, pageSize int) ([]ArticleListResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	feedKey := fmt.Sprintf("feed:user:%d", userID)

	// 1. 先获取 Feed 流总数
	total, err := s.redisClient.ZCard(ctx, feedKey).Result()
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []ArticleListResponse{}, 0, nil
	}

	// 2. 分页计算：ZREVRANGEBYSCORE 的 offset 从 0 开始
	start := int64((page - 1) * pageSize)
	if start >= total {
		return []ArticleListResponse{}, total, nil
	}

	// 3. 使用 ZREVRANGEBYSCORE 按分数（时间戳）从大到小取出文章 ID
	//    "+inf" 表示最大分数，"-inf" 表示最小分数
	//    WithScore=false，我们只需要 member（文章 ID）
	articleIDs, err := s.redisClient.ZRevRangeByScore(ctx, feedKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: start,
		Count:  int64(pageSize),
	}).Result()
	if err != nil {
		return nil, 0, err
	}

	if len(articleIDs) == 0 {
		return []ArticleListResponse{}, total, nil
	}

	// 4. 将字符串 ID 转为 uint 数组
	ids := make([]uint, 0, len(articleIDs))
	for _, idStr := range articleIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, uint(id))
	}

	if len(ids) == 0 {
		return []ArticleListResponse{}, total, nil
	}

	// 5. 从 MySQL 批量查询文章详情（使用现有的 repo 方法）
	articles, err := s.articleRepo.GetArticlesByIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	// 6. 按照 Redis 中的顺序重新排列（ZREVRANGEBYSCORE 返回的顺序）
	articleMap := make(map[uint]models.Article, len(articles))
	for _, a := range articles {
		articleMap[a.ID] = a
	}

	result := make([]ArticleListResponse, 0, len(ids))
	for _, id := range ids {
		a, ok := articleMap[id]
		if !ok {
			continue
		}

		var tagNames []string
		for _, t := range a.Tags {
			tagNames = append(tagNames, t.Name)
		}
		var categoryName string
		if a.Category != nil {
			categoryName = a.Category.Name
		}

		result = append(result, ArticleListResponse{
			ID:           a.ID,
			Title:        a.Title,
			Preview:      a.Preview,
			Likes:        a.Likes,
			Views:        a.Views,
			UserID:       a.UserID,
			CategoryName: categoryName,
			Tags:         tagNames,
			CoverImg:     a.CoverImg,
			CreatedAt:    a.CreatedAt,
		})
	}

	return result, total, nil
}

// 确保服务在编译时满足接口需求
var _ FeedService = (*feedServiceImpl)(nil)
