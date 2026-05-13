package service

import (
	"context"
	"encoding/json"
	"errors"
	"gggvrm/models"
	"gggvrm/repository"
	"log"

	"gggvrm/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type TagsService interface {
	CreateTag(ctx context.Context, tag *models.Tag) error
	GetTags(ctx context.Context, tags *[]models.Tag) error
	DelTag(ctx context.Context, id uint) error
}

type tagsServiceImpl struct {
	tagsRepo    repository.TagsRepository
	redisClient *redis.Client
	sfGroup     singleflight.Group
}

func NewTagsService(tagsRepo repository.TagsRepository, redisClient *redis.Client) TagsService {
	return &tagsServiceImpl{
		tagsRepo:    tagsRepo,
		redisClient: redisClient,
	}
}

func (s *tagsServiceImpl) CreateTag(ctx context.Context, tag *models.Tag) error {
	newName := utils.FilterSymbolsFast(tag.Name)
	if newName == "" {
		return errors.New("标签名称为空")
	}
	tag.Name = newName
	return s.tagsRepo.CreateTag(ctx, tag)
}

func (s *tagsServiceImpl) GetTags(ctx context.Context, tags *[]models.Tag) error {
	cacheData, err := s.redisClient.Get(ctx, "Tags").Result()
	if err == nil {
		if jsonErr := json.Unmarshal([]byte(cacheData), &tags); jsonErr == nil {
			return nil // 正常命中缓存
		}
		log.Println("缓存数据解析失败，尝试降级查库:", err)
	} else if !errors.Is(err, redis.Nil) {
		// 如果不是缓存未命中，而是 Redis 真的出错了（比如网络异常），记录错误日志
		log.Printf("Redis 读取异常 (key: %s): %v\n", "Tags", err)
		// 视你的业务容忍度，这里可以选择直接 return nil, err 保护 DB，或者继续向下走降级查库
	}

	// 2. 缓存未命中或解析失败，使用 Singleflight 进行查库拦截
	_, err, _ = s.sfGroup.Do("Tags", func() (interface{}, error) {

		// 3. 执行真正的数据库查询
		dbErr := s.tagsRepo.GetTags(ctx, tags)
		if dbErr != nil {
			// 修正1：查库失败必须把错误返回，绝对不能写缓存，也不能 return nil
			return nil, dbErr
		}

		// 4. 查库成功后，回写缓存
		if setErr := utils.Setcache(ctx, "Tags", tags); setErr != nil {
			log.Printf("【警告】设置Tags缓存失败 cacheKey: %s, err: %v\n", "Tags", setErr)
		}

		return tags, nil
	})

	// 5. 统一处理 Singleflight 闭包中返回的错误
	if err != nil {
		return err // 把 DB 的真实错误抛给 Controller 层
	}
	return nil
}

func (s *tagsServiceImpl) DelTag(ctx context.Context, id uint) error {
	return s.tagsRepo.DelTag(ctx, id)
}
