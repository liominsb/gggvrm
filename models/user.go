package models

import (
	"time"

	"gorm.io/gorm"
)

//带和不带[]的foreignKey的语义是不一样的，
//User表（带 []）：
//「你们这些文章，用你们身上的 UserID 来找我」
//→ 看对方的字段
//Article表（不带 []）：
//「我身上的 UserID 是用来找我的作者」
//→ 看自己的字段

type User struct {
	gorm.Model
	Username  string    `gorm:"unique" json:"username" binding:"required"`
	Password  string    `json:"-"`
	Articles  []Article `gorm:"foreignKey:UserID" json:"articles"`
	Favorites []Article `gorm:"many2many:user_article_favor;"`
}

type UserArticleFavor struct {
	UserId    uint `gorm:"primaryKey"`
	ArticleID uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
