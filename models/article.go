package models // Package models 模型

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name string `json:"name"` // 分类名称
}
type Tag struct {
	gorm.Model
	Name     string    `json:"name" gorm:"unique"` // 标签名称
	Articles []Article `gorm:"many2many:article_tags;"`
}

type Article struct {
	gorm.Model
	Title    string    `json:"title" binding:"required"`   //标题
	Content  string    `json:"content" binding:"required"` //内容
	Preview  string    `json:"preview" binding:"required"` //预览
	Likes    int       `json:"likes" gorm:"default:0"`     //点赞数，默认为0
	Views    int       `json:"views" gorm:"default:0"`     //浏览数，默认为0
	User     *User     `json:"user" gorm:"foreignKey:UserID"`
	UserID   uint      `json:"user_id" binding:"required"`
	CoverImg string    `json:"cover_img"`                                                         //【新增】封面图的 URL
	Comments []Comment `json:"comments" gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE;"` //评论

	CategoryID uint      `json:"category_id"`                           // 记录分类的 ID
	Category   *Category `json:"category" gorm:"foreignKey:CategoryID"` //类别

	Tags []Tag `json:"tags" gorm:"many2many:article_tags;"`

	FavoredBy []User `gorm:"many2many:user_article_favor"` //被哪些用户收藏
}

type ArticleTags struct {
	TagID     uint `gorm:"primaryKey"`
	ArticleID uint `gorm:"primaryKey"`
	CreatedAt time.Time
}

type Comment struct {
	gorm.Model
	ArticleID uint   `json:"article_id" binding:"required"` // 外键，关联到 Article.ID
	UserID    uint   `json:"user_id" binding:"required"`    // 评论者 ID
	Content   string `json:"content" binding:"required"`    // 评论具体内容
}
