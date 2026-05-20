package repository

import (
	"context"
	"errors"
	"gggvrm/models"

	"gorm.io/gorm"
)

type LikeRepository interface {
	GetArticleLikes(ctx context.Context, articleID uint) (int, error)
	GetArticleByIDWithPreload(ctx context.Context, article *models.Article, id string) error
}

type likeRepoImpl struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepoImpl{
		db: db,
	}
}

func (r *likeRepoImpl) GetArticleLikes(ctx context.Context, articleID uint) (int, error) {
	var article models.Article
	if err := r.db.WithContext(ctx).Select("likes").Where("id = ?", articleID).First(&article).Error; err != nil {
		// 查不到文章
		return 0, errors.New("文章不存在")
	}

	return article.Likes, nil
}

func (r *likeRepoImpl) GetArticleByIDWithPreload(ctx context.Context, article *models.Article, id string) error {
	return r.db.WithContext(ctx).Preload("Category").Preload("Tags").Where("id = ?", id).First(article).Error
}
