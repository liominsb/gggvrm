package service

import (
	"context"
	"errors"
	"gggvrm/models"
	"gggvrm/repository"

	"gggvrm/utils"
	"log"
	"time"

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

	cacheKey := "Tags"

	// 第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("【警告】第一次删除缓存失败: %s, err: %v\n", cacheKey, err)
	}

	if err := s.tagsRepo.CreateTag(ctx, tag); err != nil {
		return err
	}

	go func() {
		time.Sleep(100 * time.Millisecond) //延时

		// 第二次删除缓存
		if err := s.redisClient.Del(context.Background(), cacheKey).Err(); err != nil {
			log.Printf("【警告】延时双删失败 cacheKey: %s, err: %v\n", cacheKey, err)
			return
		}
	}()

	return nil
}

func (s *tagsServiceImpl) GetTags(ctx context.Context, tags *[]models.Tag) error {
	// 使用 Singleflight 防止缓存击穿，内部使用 GetCacheOrQuery 处理缓存逻辑
	v, err, _ := s.sfGroup.Do("Tags", func() (interface{}, error) {
		result, err := utils.GetCacheOrQuery(ctx, s.redisClient, "Tags", func() (*[]models.Tag, error) {
			var t []models.Tag
			if err := s.tagsRepo.GetTags(ctx, &t); err != nil {
				return nil, err
			}
			return &t, nil
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	})

	if err != nil {
		return err
	}

	*tags = *v.(*[]models.Tag)
	return nil
}

func (s *tagsServiceImpl) DelTag(ctx context.Context, id uint) error {
	cacheKey := "Tags"

	// 第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		log.Printf("【警告】第一次删除缓存失败: %s, err: %v\n", cacheKey, err)
	}

	if err := s.tagsRepo.DelTag(ctx, id); err != nil {
		return err
	}

	go func() {
		time.Sleep(100 * time.Millisecond) //延时

		// 第二次删除缓存
		if err := s.redisClient.Del(context.Background(), cacheKey).Err(); err != nil {
			log.Printf("【警告】延时双删失败 cacheKey: %s, err: %v\n", cacheKey, err)
		}
	}()

	return nil
}
