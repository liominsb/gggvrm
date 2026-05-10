package repository

import (
	"context"
	"gggvrm/models"
	"log"

	"gorm.io/gorm"
)

type CommentsRepository interface {
	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, id uint) (*models.Comment, error)
	GetArticleByID(ctx context.Context, id uint) (*models.Article, error)
	DelComment(ctx context.Context, comment *models.Comment) error
	GetComments(ctx context.Context, articleID uint) ([]models.Comment, error)
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

func (r *commentsRepoImpl) GetCommentByID(ctx context.Context, id uint) (*models.Comment, error) {
	comment := models.Comment{}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		log.Println("数据库查询评论失败:", err)
		return nil, err
	}
	return &comment, nil
}

func (r *commentsRepoImpl) GetArticleByID(ctx context.Context, id uint) (*models.Article, error) {
	article := models.Article{}
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&article).Error; err != nil {
		log.Println("数据库查询文章失败:", err)
		return nil, err
	}
	return &article, nil
}

func (r *commentsRepoImpl) DelComment(ctx context.Context, comment *models.Comment) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(&comment).Error; err != nil {
		log.Println("数据库删除评论失败:", err)
		return err
	}
	return nil
}

func (r *commentsRepoImpl) GetComments(ctx context.Context, articleID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.db.WithContext(ctx).Where("article_id = ?", articleID).Find(&comments).Error; err != nil {
		log.Println("数据库查询评论失败:", err)
		return comments, err
	}
	return comments, nil
}
