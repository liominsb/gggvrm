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
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// ArticleListResponse 定义文章列表的轻量级返回结构（不包含 Content）
type ArticleListResponse struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Preview      string    `json:"preview"`
	Likes        int       `json:"likes"`
	Views        int       `json:"views"`
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`      // 附加作者用户名
	CategoryName string    `json:"category_name"` // 附加分类名
	Tags         []string  `json:"tags"`          // 附加标签数组
	CoverImg     string    `json:"cover_img"`     // 封面图的 URL
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
}

// ArticleCache 专门用来打包存入 Redis 的结构
type ArticleCache struct {
	Total int64                 `json:"total"`
	Data  []ArticleListResponse `json:"data"`
}

type ArticleService interface {
	CreateArticle(ctx context.Context, article *models.Article, tagIDs []uint) error
	GetArticles(ctx context.Context, page, pageSize, categoryID, tagID int, keyword string, userID uint) (*ArticleCache, int, int, int64, error)
	GetArticlesByID(ctx context.Context, id string) (*models.Article, []models.Comment, string, error)
	DelArticle(ctx context.Context, articleID string, userID uint) error
	UpdateArticle(ctx context.Context, articleID string, userID uint, input struct {
		Title      string
		Content    string
		Preview    string
		CategoryID uint
		TagIDs     []uint
		CoverImg   string
	}) (*models.Article, error)
	GetArticlesByCursor(ctx context.Context, cursor uint64, limit int) ([]models.Article, uint, bool, error)
}

type articleServiceImpl struct {
	articleRepo repository.ArticleRepository
	commentRepo repository.CommentsRepository
	redisClient *redis.Client
}

func NewArticleService(articleRepo repository.ArticleRepository, commentRepo repository.CommentsRepository, redisClient *redis.Client) ArticleService {
	return &articleServiceImpl{articleRepo: articleRepo, commentRepo: commentRepo, redisClient: redisClient}
}

// 辅助函数：清理所有文章分页列表的缓存
func (s *articleServiceImpl) clearArticlesCache(ctx context.Context) {
	var cursor uint64 = 0
	var count int64 = 100
	var keys []string
	var err error
	for {
		keys, cursor, err = s.redisClient.Scan(ctx, cursor, "articles:page:*", count).Result()
		if err != nil {
			log.Println(err)
			break
		}
		if len(keys) > 0 {
			err = s.redisClient.Del(ctx, keys...).Err()
			if err != nil {
				log.Println(err)
			}
		}
		if cursor == 0 {
			break
		}
	}
}

func (s *articleServiceImpl) CreateArticle(ctx context.Context, article *models.Article, tagIDs []uint) error {
	if len(tagIDs) > 0 {
		var tags []models.Tag
		if err := s.articleRepo.GetTagsByIDs(ctx, &tags, tagIDs); err != nil {
			return err
		}
		article.Tags = tags
	}

	if err := s.articleRepo.CreateArticle(ctx, article); err != nil {
		return err
	}

	s.clearArticlesCache(ctx)

	// 通过 MQ 异步推送 Feed 流给作者的所有粉丝
	msgData, _ := json.Marshal(map[string]interface{}{
		"action":     "create_article",
		"article_id": article.ID,
		"user_id":    article.UserID,
		"timestamp":  time.Now().Unix(),
	})
	if err := mq.PublishMessage("article_tasks", msgData); err != nil {
		log.Printf("【RabbitMQ警告】发送文章消息失败: %v", err)
	}

	return nil
}

func (s *articleServiceImpl) GetArticles(ctx context.Context, page, pageSize, categoryID, tagID int, keyword string, userID uint) (*ArticleCache, int, int, int64, error) {
	if page <= 0 {
		page = 1
	}
	if page > 1000 {
		return nil, 0, 0, 0, errors.New("请求页码超出最大支持范围")
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	dynamicCacheKey := fmt.Sprintf("articles:page:%d:size:%d:cat:%d:tag:%d:user:%d", page, pageSize, categoryID, tagID, userID)

	cacheObj, err := utils.GetCacheOrQuery(ctx, s.redisClient, dynamicCacheKey, func() (*ArticleCache, error) {
		offset := (page - 1) * pageSize
		var articles []models.Article
		total, err := s.articleRepo.GetArticlesWithPagination(ctx, &articles, offset, pageSize, categoryID, tagID, keyword, userID)
		if err != nil {
			return nil, err
		}

		// 即使结果为空，也缓存（防止缓存穿透）
		if total == 0 {
			cacheObj := ArticleCache{Total: 0, Data: []ArticleListResponse{}}
			return &cacheObj, nil
		}

		// 数据转换 (DTO 映射)
		var response = make([]ArticleListResponse, 0)
		for _, a := range articles {
			var tagNames []string
			for _, t := range a.Tags {
				tagNames = append(tagNames, t.Name)
			}
			var categoryName string
			if a.Category != nil {
				categoryName = a.Category.Name
			}
			var username string
			if a.User != nil {
				username = a.User.Username
			}
			response = append(response, ArticleListResponse{
				ID:           a.ID,
				Title:        a.Title,
				Preview:      a.Preview,
				Likes:        a.Likes,
				Views:        a.Views,
				UserID:       a.UserID,
				Username:     username,
				CategoryName: categoryName,
				Tags:         tagNames,
				CoverImg:     a.CoverImg,
				CreatedAt:    a.CreatedAt,
			})
		}

		cacheObj := ArticleCache{
			Total: total,
			Data:  response,
		}
		return &cacheObj, nil
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if cacheObj == nil {
		return nil, 0, 0, 0, err
	}
	return cacheObj, page, pageSize, cacheObj.Total, nil
}

func (s *articleServiceImpl) GetArticlesByID(ctx context.Context, id string) (*models.Article, []models.Comment, string, error) {
	var article models.Article
	var comments []models.Comment
	var likes string

	eg, gCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		cacheKey := fmt.Sprintf("article:detail:%s", id)
		result, err := utils.GetCacheOrQuery(gCtx, s.redisClient, cacheKey, func() (*models.Article, error) {
			if err := s.articleRepo.GetArticleByIDWithPreload(gCtx, &article, id); err != nil {
				return nil, err
			}
			return &article, nil
		})
		if err != nil {
			return err
		}
		article = *result
		return nil
	})

	eg.Go(func() error {
		articleID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return err
		}
		cacheKey := fmt.Sprintf("article:%d:comments", articleID)
		result, err := utils.GetCacheOrQuery(gCtx, s.redisClient, cacheKey, func() (*[]models.Comment, error) {
			comments, err := s.commentRepo.GetComments(gCtx, uint(articleID))
			if err != nil {
				return nil, err
			}
			return &comments, nil
		})
		if err != nil {
			return err
		}
		comments = *result
		return nil
	})

	eg.Go(func() error {
		var err error
		var temp models.Article

		likeKey := "article:" + id + ":likes"

		likes, err = s.redisClient.Get(gCtx, likeKey).Result()

		if errors.Is(err, redis.Nil) {
			// 只查询 likes 字段，提高效率
			if err := s.articleRepo.GetArticleByID(gCtx, &temp, id); err != nil {
				return err
			}

			if err := s.redisClient.SetNX(gCtx, likeKey, temp.Likes, 0).Err(); err != nil {
				return err
			}
			likes = strconv.Itoa(temp.Likes)
		} else if err != nil {
			return err
		}

		return nil
	})

	go func() {
		if err := s.redisClient.Incr(ctx, fmt.Sprintf("article:%s:views", id)).Err(); err != nil {
			fmt.Printf("Redis警告 增加文章浏览量失败: %v\n", err)
			return
		}
	}()

	if err := eg.Wait(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, "", errors.New("该文章不存在")
		}
		return nil, nil, "", err
	}

	return &article, comments, likes, nil
}

func (s *articleServiceImpl) DelArticle(ctx context.Context, articleID string, userID uint) error {
	var article models.Article

	if err := s.articleRepo.GetArticleByID(ctx, &article, articleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文章不存在")
		}
		return err
	}

	if article.UserID != userID {
		return errors.New("无权限删除该文章")
	}

	// 第一次删除缓存
	s.clearArticlesCache(ctx)
	s.redisClient.Del(ctx, fmt.Sprintf("article:detail:%s", articleID))
	s.redisClient.Del(ctx, fmt.Sprintf("article:%s:comments", articleID))
	s.redisClient.Del(ctx, fmt.Sprintf("article:%s:likes", articleID))

	if err := s.articleRepo.DeleteArticle(ctx, &article); err != nil {
		return err
	}

	wd, _ := os.Getwd()
	fullPath := filepath.Join(wd, article.CoverImg)
	if err := os.Remove(fullPath); err != nil {
		log.Printf("删除文件失败: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond) //延时

		// 第二次删除缓存
		s.clearArticlesCache(context.Background())
		s.redisClient.Del(context.Background(), fmt.Sprintf("article:detail:%s", articleID))
		s.redisClient.Del(context.Background(), fmt.Sprintf("article:%s:comments", articleID))
		s.redisClient.Del(context.Background(), fmt.Sprintf("article:%s:likes", articleID))
	}()

	return nil
}

func (s *articleServiceImpl) UpdateArticle(ctx context.Context, articleID string, userID uint, input struct {
	Title      string
	Content    string
	Preview    string
	CategoryID uint
	TagIDs     []uint
	CoverImg   string
}) (*models.Article, error) {
	var article models.Article

	// 1. 查找文章是否存在
	if err := s.articleRepo.GetArticleByID(ctx, &article, articleID); err != nil {
		return nil, errors.New("文章不存在")
	}

	// 2. 越权校验：只能修改自己的文章
	if article.UserID != userID {
		return nil, errors.New("无权修改他人的文章")
	}

	// 3. 处理标签
	if input.TagIDs != nil {
		if len(input.TagIDs) > 0 {
			var tags []models.Tag
			if err := s.articleRepo.GetTagsByIDs(ctx, &tags, input.TagIDs); err != nil {
				return nil, errors.New("查询标签失败")
			}
			if err := s.articleRepo.ReplaceArticleTags(ctx, &article, tags); err != nil {
				return nil, errors.New("更新标签失败")
			}
		} else {
			if err := s.articleRepo.ClearArticleTags(ctx, &article); err != nil {
				return nil, errors.New("清空标签失败")
			}
		}
	}

	// 第一次删除缓存
	s.clearArticlesCache(ctx)
	s.redisClient.Del(ctx, fmt.Sprintf("article:detail:%s", articleID))

	updateData := map[string]interface{}{
		"title":       input.Title,
		"content":     input.Content,
		"preview":     input.Preview,
		"category_id": input.CategoryID,
		"cover_img":   input.CoverImg,
	}

	if err := s.articleRepo.UpdateArticle(ctx, &article, updateData); err != nil {
		return nil, err
	}

	go func() {
		time.Sleep(100 * time.Millisecond) //延时

		// 第二次删除缓存
		s.clearArticlesCache(ctx)
		s.redisClient.Del(ctx, fmt.Sprintf("article:detail:%s", articleID))
	}()

	return &article, nil
}

func (s *articleServiceImpl) GetArticlesByCursor(ctx context.Context, cursor uint64, limit int) ([]models.Article, uint, bool, error) {
	var articles []models.Article

	if err := s.articleRepo.GetArticlesByCursor(ctx, &articles, cursor, limit); err != nil {
		return nil, 0, false, err
	}

	hasMore := len(articles) > limit
	if hasMore {
		articles = articles[:limit] // 去掉多查的那条
	}

	nextCursor := articles[len(articles)-1].ID

	return articles, nextCursor, hasMore, nil
}
