package repository

import (
	"context"
	"gggvrm/models"
	"log"

	"gorm.io/gorm"
)

type CommentsRepository interface {
	CreateComment(ctx context.Context, comment *models.Comment) error
}

type commentsRepoImpl struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentsRepository {
	return &commentsRepoImpl{db: db}
}

func (r *commentsRepoImpl) CreateComment(ctx context.Context, comment *models.Comment) error {

	if err := r.db.WithContext(ctx).Create(&comment).Error; err != nil {
		log.Println("数据库插入评论失败:", err)
		return err
	}
	return nil
}
