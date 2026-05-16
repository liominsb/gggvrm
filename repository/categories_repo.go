package repository

import (
	"context"
	"gggvrm/models"

	"gorm.io/gorm"
)

type CateRepository interface {
	CreateCate(ctx context.Context, cate *models.Category) error
	GetCateByID(ctx context.Context, categoryID uint) (*models.Category, error)
	DeleteCate(ctx context.Context, categoryID uint) error
	GetCates(ctx context.Context, cates *[]models.Category) error
}

type cateRepoImpl struct {
	db *gorm.DB
}

func NewCateRepository(db *gorm.DB) CateRepository {
	return &cateRepoImpl{
		db: db,
	}
}

func (c *cateRepoImpl) CreateCate(ctx context.Context, cate *models.Category) error {
	return c.db.WithContext(ctx).Create(cate).Error
}

func (c *cateRepoImpl) GetCateByID(ctx context.Context, categoryID uint) (*models.Category, error) {
	var cate *models.Category
	if err := c.db.WithContext(ctx).First(&cate, "id = ?", categoryID).Error; err != nil {
		return nil, err
	}
	return cate, nil
}

func (c *cateRepoImpl) DeleteCate(ctx context.Context, categoryID uint) error {
	return c.db.WithContext(ctx).Unscoped().Delete(&models.Category{}, "id = ?", categoryID).Error
}

func (c *cateRepoImpl) GetCates(ctx context.Context, cates *[]models.Category) error {
	return c.db.WithContext(ctx).Find(cates).Error
}
