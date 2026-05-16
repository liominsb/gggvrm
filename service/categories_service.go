package service

import (
	"context"
	"fmt"
	"gggvrm/models"
	"gggvrm/repository"
	"gggvrm/utils"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type CateService interface {
	CreateCate(ctx context.Context, cate *models.Category) error
	GetCates(ctx context.Context, cates *[]models.Category) ([]models.Category, error)
	DeleteCate(ctx context.Context, cateID uint) error
	GetCateByID(ctx context.Context, cateID uint) (*models.Category, error)
}

type cateServiceImpl struct {
	cateRepo    repository.CateRepository
	redisClient *redis.Client
	sfGroup     singleflight.Group
}

func NewCateService(cateRepo repository.CateRepository, redisClient *redis.Client) CateService {
	return &cateServiceImpl{
		cateRepo:    cateRepo,
		redisClient: redisClient,
	}
}

func (s *cateServiceImpl) CreateCate(ctx context.Context, cate *models.Category) error {
	cacheKey := "categories:all"
	//第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		fmt.Printf("【Redis警告】清理分类 %d 缓存失败: %v\n", cate.ID, err)
	}

	err := s.cateRepo.CreateCate(ctx, cate)
	if err != nil {
		log.Println("数据库删除分类失败:", err)
		return fmt.Errorf("删除分类失败: %w", err)
	}

	go func(bgctx context.Context) {
		time.Sleep(100 * time.Millisecond) //延时
		//第二次删除缓存
		if err := s.redisClient.Del(bgctx, cacheKey).Err(); err != nil {
			fmt.Printf("【Redis警告】清理分类 %d 缓存失败: %v\n", cate.ID, err)
		}
	}(context.Background())
	return nil
}

func (s *cateServiceImpl) GetCates(ctx context.Context, cates *[]models.Category) ([]models.Category, error) {
	cacheKey := "categories:all"
	// 使用 Singleflight 防止缓存击穿，内部使用 GetCacheOrQuery 处理缓存逻辑
	v, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*[]models.Category, error) {
			err := s.cateRepo.GetCates(ctx, cates)
			if err != nil {
				return nil, err
			}
			return cates, nil
		})
		if err != nil {
			return nil, err
		}
		return *result, nil
	})

	if err != nil {
		return nil, err
	}

	return v.([]models.Category), nil
}

func (s *cateServiceImpl) DeleteCate(ctx context.Context, cateID uint) error {
	cacheKey := fmt.Sprintf("category:%d", cateID)
	//第一次删除缓存
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		fmt.Printf("【Redis警告】清理分类 %d 缓存失败: %v\n", cateID, err)
	}

	err := s.cateRepo.DeleteCate(ctx, cateID)
	if err != nil {
		log.Println("数据库删除分类失败:", err)
		return fmt.Errorf("删除分类失败: %w", err)
	}

	go func(bgctx context.Context) {
		time.Sleep(100 * time.Millisecond) //延时
		//第二次删除缓存
		if err := s.redisClient.Del(bgctx, cacheKey).Err(); err != nil {
			fmt.Printf("【Redis警告】清理分类 %d 缓存失败: %v\n", cateID, err)
		}
	}(context.Background())
	return nil
}

// 意义不明
func (s *cateServiceImpl) GetCateByID(ctx context.Context, cateID uint) (*models.Category, error) {
	cacheKey := fmt.Sprintf("category:%d", cateID)
	// 使用 Singleflight 防止缓存击穿，内部使用 GetCacheOrQuery 处理缓存逻辑
	v, err, _ := s.sfGroup.Do(cacheKey, func() (interface{}, error) {
		result, err := utils.GetCacheOrQuery(ctx, s.redisClient, cacheKey, func() (*models.Category, error) {
			cate, err := s.cateRepo.GetCateByID(ctx, cateID)
			if err != nil {
				return nil, err
			}
			return cate, nil
		})
		if err != nil {
			return nil, err
		}
		return *result, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*models.Category), nil
}
