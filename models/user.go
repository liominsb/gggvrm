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
	Username  string    `gorm:"size:50;uniqueIndex" json:"username" binding:"required"`
	Password  string    `json:"-"`
	Articles  []Article `gorm:"foreignKey:UserID" json:"articles"`
	Favorites []Article `gorm:"many2many:user_article_favor;"` //收藏夹

	// Followings: 我关注的人 (我是 Follower，目标是 Followee)
	// joinForeignKey 指向中间表里代表“我”的字段，joinReferences 指向代表“对方”的字段
	Followings []User `gorm:"many2many:user_follows;joinForeignKey:FollowerID;joinReferences:FolloweeID" json:"followings"`

	// Followers: 关注我的人 (我是 Followee，目标是 Follower)
	Followers []User `gorm:"many2many:user_follows;joinForeignKey:FolloweeID;joinReferences:FollowerID" json:"followers"`
}

// UserFollow 关注中间表，记录谁关注了谁，以及关注的时间
type UserFollow struct {
	FollowerID uint      `gorm:"primaryKey;index:idx_follower_followee,unique"` // 关注者 ID (谁发起的关注)
	FolloweeID uint      `gorm:"primaryKey;index:idx_follower_followee,unique"` // 被关注者 ID (目标大V)
	CreatedAt  time.Time // 关注时间
}

// 收藏中间表，记录哪个用户收藏了哪个文章，以及收藏的时间
type UserArticleFavor struct {
	UserID    uint `gorm:"primaryKey"`
	ArticleID uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
