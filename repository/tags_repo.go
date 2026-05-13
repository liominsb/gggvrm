package repository

import (
	"context"
	"gggvrm/models"

	"gorm.io/gorm"
)

type TagsRepository interface {
	CreateTag(ctx context.Context, tag *models.Tag) error
	GetTags(ctx context.Context, tags *[]models.Tag) error
	DelTag(ctx context.Context, id uint) error
}

type tagsRepoImpl struct {
	db *gorm.DB
}

func NewTagsRepository(db *gorm.DB) TagsRepository {
	return &tagsRepoImpl{
		db: db,
	}
}

func (r *tagsRepoImpl) CreateTag(ctx context.Context, tag *models.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagsRepoImpl) GetTags(ctx context.Context, tags *[]models.Tag) error {
	return r.db.WithContext(ctx).Find(tags).Error

}

func (r *tagsRepoImpl) DelTag(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", id).Delete(&models.Tag{}).Error
}
