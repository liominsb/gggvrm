package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string    `gorm:"size:50;uniqueIndex" json:"username" binding:"required"`
	Password  string    `json:"-"`
	Bio       string    `gorm:"size:500" json:"bio"`
	Image     string    `gorm:"size:500" json:"image"`
	Articles  []Article `gorm:"foreignKey:UserID" json:"articles"`
	Favorites []Article `gorm:"many2many:user_article_favors;"`

	// Followings are users this user follows.
	Followings []User `gorm:"many2many:user_follows;joinForeignKey:FollowerID;joinReferences:FolloweeID" json:"followings"`

	// Followers are users who follow this user.
	Followers []User `gorm:"many2many:user_follows;joinForeignKey:FolloweeID;joinReferences:FollowerID" json:"followers"`
}

type UserFollow struct {
	FollowerID uint `gorm:"primaryKey;index:idx_follower_followee,unique"`
	FolloweeID uint `gorm:"primaryKey;index:idx_follower_followee,unique"`
	CreatedAt  time.Time
}

type UserArticleFavor struct {
	UserID    uint `gorm:"primaryKey"`
	ArticleID uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
