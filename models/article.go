package models // Package models 模型

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name string `json:"name"` // 分类名称
}
type Tag struct {
	gorm.Model
	Name string `json:"name"` // 标签名称
}

// Article 表示一篇文章
// Title  文章标题（必填）
// Content 文章正文（必填）
// Preview 文章预览（必填）
// Likes   点赞数，默认为 0
// Comments 文章评论
type Article struct {
	gorm.Model
	Title    string    `json:"title" binding:"required"`   //标题
	Content  string    `json:"content" binding:"required"` //内容
	Preview  string    `json:"preview" binding:"required"` //预览
	Likes    int       `json:"likes" gorm:"default:0"`     //喜好
	User     User      `json:"user" gorm:"foreignKey:UserID"`
	UserID   uint      `json:"user_id" binding:"required"`
	Comments []Comment `json:"comments" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE;"` //评论

	CategoryID uint     `json:"category_id"` // 记录分类的 ID
	Category   Category `json:"category" gorm:"foreignKey:CategoryID"`

	Tags []Tag `json:"tags" gorm:"many2many:article_tags;"`
}

type Comment struct {
	gorm.Model
	ArticleID uint   `json:"article_id" binding:"required"` // 外键，关联到 Article.ID
	UserID    uint   `json:"user_id" binding:"required"`    // 评论者 ID
	Content   string `json:"content" binding:"required"`    // 评论具体内容
}
