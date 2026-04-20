package config

import (
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func initDB() {
	dsn := Appconf.Database.Dsn
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()

	sqlDB.SetMaxIdleConns(Appconf.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(Appconf.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	global.Db = db

	// Tag 的 Articles 字段使用 TagArticles 中间表
	err = global.Db.SetupJoinTable(&models.Tag{}, "Articles", &models.ArticleTags{})
	if err != nil {
		fmt.Println("注册 Tag 的中间表失败:", err)
		return
	}
	// Article 的 Tags 字段使用 TagArticles 中间表
	err = global.Db.SetupJoinTable(&models.Article{}, "Tags", &models.ArticleTags{})
	if err != nil {
		fmt.Println("注册 Article 的中间表失败:", err)
		return
	}

	err = global.Db.SetupJoinTable(&models.User{}, "Favorites", &models.UserArticleFavor{})
	if err != nil {
		fmt.Println("注册 User 的中间表失败:", err)
		return
	}

	err = global.Db.SetupJoinTable(&models.Article{}, "FavoredBy", &models.UserArticleFavor{})
	if err != nil {
		fmt.Println("注册 Article 的中间表失败:", err)
		return
	}

	err = global.Db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.Comment{},
		&models.Tag{},
		&models.Category{},
	)
	if err != nil {
		log.Fatalf("数据库表结构同步失败: %v", err)
	}
	log.Println("数据库表结构同步成功！")
}
