package repository

import (
	"context"
	"gggvrm/models"

	"gorm.io/gorm"
)

// FavoriteRepository 收藏数据访问接口
type FavoriteRepository interface {
	AddFavorite(ctx context.Context, userID, articleID uint) error                                            // 添加收藏
	RemoveFavorite(ctx context.Context, userID, articleID uint) error                                         // 取消收藏
	IsFavorited(ctx context.Context, userID, articleID uint) (bool, error)                                    // 查询用户是否已收藏该文章
	GetFavoriteCount(ctx context.Context, articleID uint) (int64, error)                                      // 获取文章的收藏总数
	GetUserFavorites(ctx context.Context, userID uint, offset, pageSize int) ([]models.Article, int64, error) // 分页获取用户的收藏列表
}

type favoriteRepoImpl struct {
	db *gorm.DB
}

// NewFavoriteRepository 创建收藏仓库实例
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepoImpl{db: db}
}

// AddFavorite 向 user_article_favor 表插入一条收藏记录
func (r *favoriteRepoImpl) AddFavorite(ctx context.Context, userID, articleID uint) error {
	favor := models.UserArticleFavor{
		UserID:    userID,
		ArticleID: articleID,
	}
	return r.db.WithContext(ctx).Create(&favor).Error
}

// RemoveFavorite 从 user_article_favor 表删除一条收藏记录
func (r *favoriteRepoImpl) RemoveFavorite(ctx context.Context, userID, articleID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&models.UserArticleFavor{}).Error
}

// IsFavorited 检查用户是否已收藏指定文章
func (r *favoriteRepoImpl) IsFavorited(ctx context.Context, userID, articleID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserArticleFavor{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count).Error
	return count > 0, err
}

// GetFavoriteCount 统计指定文章的收藏总数
func (r *favoriteRepoImpl) GetFavoriteCount(ctx context.Context, articleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserArticleFavor{}).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, err
}

// GetUserFavorites 通过中间表关联查询用户收藏的文章列表，支持分页
func (r *favoriteRepoImpl) GetUserFavorites(ctx context.Context, userID uint, offset, pageSize int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	// 通过中间表关联查询用户收藏的文章
	query := r.db.WithContext(ctx).
		Joins("JOIN user_article_favor ON user_article_favor.article_id = articles.id").
		Where("user_article_favor.user_id = ?", userID)

	if err := query.Model(&models.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 初始化为空切片，防止返回 null
	articles = make([]models.Article, 0)

	if total > 0 && int64(offset) < total {
		if err := query.
			Preload("Category").Preload("Tags").
			Omit("Content").
			Order("user_article_favor.created_at DESC").
			Limit(pageSize).Offset(offset).
			Find(&articles).Error; err != nil {
			return nil, 0, err
		}
	}

	return articles, total, nil
}
