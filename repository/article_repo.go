package repository

import (
	"context"
	"gggvrm/models"

	"gorm.io/gorm"
)

type ArticleRepository interface {
	CreateArticle(ctx context.Context, article *models.Article) error
	GetArticleByID(ctx context.Context, article *models.Article, id string) error
	GetArticleByIDWithPreload(ctx context.Context, article *models.Article, id string) error
	GetArticlesByIDs(ctx context.Context, ids []uint) ([]models.Article, error)
	DeleteArticle(ctx context.Context, article *models.Article) error
	UpdateArticle(ctx context.Context, article *models.Article, updateData map[string]interface{}) error
	GetArticlesWithPagination(ctx context.Context, articles *[]models.Article, offset, pageSize int, categoryID, tagID int, keyword string) (int64, error)
	GetArticlesByCursor(ctx context.Context, articles *[]models.Article, cursor uint64, limit int) error
	GetTagsByIDs(ctx context.Context, tags *[]models.Tag, ids []uint) error
	ReplaceArticleTags(ctx context.Context, article *models.Article, tags []models.Tag) error
	ClearArticleTags(ctx context.Context, article *models.Article) error
}

type articleRepoImpl struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepoImpl{db: db}
}

func (r *articleRepoImpl) CreateArticle(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *articleRepoImpl) GetArticleByID(ctx context.Context, article *models.Article, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).First(article).Error
}

func (r *articleRepoImpl) GetArticleByIDWithPreload(ctx context.Context, article *models.Article, id string) error {
	return r.db.WithContext(ctx).Preload("Category").Preload("Tags").Where("id = ?", id).First(article).Error
}

func (r *articleRepoImpl) GetArticlesByIDs(ctx context.Context, ids []uint) ([]models.Article, error) {
	var articles []models.Article
	if err := r.db.WithContext(ctx).Preload("Category").Preload("Tags").
		Where("id IN ?", ids).Omit("Content").Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *articleRepoImpl) DeleteArticle(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Delete(article).Error
}

func (r *articleRepoImpl) UpdateArticle(ctx context.Context, article *models.Article, updateData map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(article).Preload("Category").Preload("Tags").Updates(updateData).Find(article).Error
}

func (r *articleRepoImpl) GetArticlesWithPagination(ctx context.Context, articles *[]models.Article, offset, pageSize int, categoryID, tagID int, keyword string) (int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Article{})

	// 分类过滤
	if categoryID > 0 {
		query = query.Where("articles.category_id = ?", categoryID)
	}

	// 标签过滤（多对多）
	if tagID > 0 {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", tagID)
	}

	// 关键词搜索
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where("articles.title LIKE ? OR articles.preview LIKE ?", likePattern, likePattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}

	// 初始化为空切片，防止返回 null
	*articles = make([]models.Article, 0)

	if total > 0 && int64(offset) < total {
		if err := query.Preload("Category").Preload("Tags").Omit("Content").
			Order("articles.id desc").Limit(pageSize).Offset(offset).Find(articles).Error; err != nil {
			return 0, err
		}
	}

	return total, nil
}

func (r *articleRepoImpl) GetArticlesByCursor(ctx context.Context, articles *[]models.Article, cursor uint64, limit int) error {
	return r.db.WithContext(ctx).Where("id < ?", cursor).
		Order("id desc").
		Limit(limit + 1). // 多查一条判断是否有下一页
		Find(articles).Error
}

func (r *articleRepoImpl) GetTagsByIDs(ctx context.Context, tags *[]models.Tag, ids []uint) error {
	return r.db.WithContext(ctx).Where("id IN ?", ids).Find(tags).Error
}

func (r *articleRepoImpl) ReplaceArticleTags(ctx context.Context, article *models.Article, tags []models.Tag) error {
	return r.db.WithContext(ctx).Model(article).Association("Tags").Replace(tags)
}

func (r *articleRepoImpl) ClearArticleTags(ctx context.Context, article *models.Article) error {
	return r.db.WithContext(ctx).Model(article).Association("Tags").Clear()
}
