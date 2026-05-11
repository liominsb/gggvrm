package utils // Package utils 实用性

import (
	"context"
	"encoding/json"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

// 检查密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Setcache 设置缓存
func Setcache(ctx context.Context, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		return err
	}
	a := time.Duration(rand.IntN(5) + 10)
	if err := global.RedisDB.Set(ctx, key, valueJSON, a*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// 同步点赞数和浏览数到数据库
func SyncSql(ctx context.Context) {
	for {
		time.Sleep(1 * time.Minute)
		var cursor uint64 // 初始游标为 0
		var cursor1 uint64
		var keys []string
		var err error
		for {
			keys, cursor, err = global.RedisDB.Scan(ctx, cursor, "article:*:likes", 100).Result()
			if err != nil {
				fmt.Println("获取 Redis Keys 失败:", err)
				break
			}

			for _, key := range keys {
				parts := strings.Split(key, ":")
				if len(parts) != 3 {
					continue
				}

				articleIDStr := parts[1]
				articleID, err := strconv.Atoi(articleIDStr)
				if err != nil {
					continue
				}

				likesStr, err := global.RedisDB.Get(ctx, key).Result()
				if err != nil {
					continue
				}
				likes, err := strconv.Atoi(likesStr)
				if err != nil {
					continue
				}

				if err := global.Db.Model(&models.Article{}).Where("id = ?", articleID).Update("likes", likes).Error; err != nil {
					fmt.Printf("更新文章 %d 点赞数失败: %v\n", articleID, err)
				}
			}

			fmt.Println("已同步点赞数到数据库")

			if cursor == 0 {
				break
			}
		}

		for {
			keys, cursor1, err = global.RedisDB.Scan(ctx, cursor1, "article:*:views", 100).Result()
			if err != nil {
				fmt.Println("获取 Redis Keys 失败:", err)
				break
			}

			for _, key := range keys {
				parts := strings.Split(key, ":")
				if len(parts) != 3 {
					continue
				}

				articleIDStr := parts[1]
				articleID, err := strconv.Atoi(articleIDStr)
				if err != nil {
					continue
				}

				viewsStr, err := global.RedisDB.Get(ctx, key).Result()
				if err != nil {
					continue
				}
				views, err := strconv.Atoi(viewsStr)
				if err != nil {
					continue
				}

				if err := global.Db.Model(&models.Article{}).Where("id = ?", articleID).Update("views", views).Error; err != nil {
					fmt.Printf("更新文章 %d 浏览数失败: %v\n", articleID, err)
				}
			}
			fmt.Println("已同步浏览数到数据库")

			if cursor1 == 0 {
				break
			}
		}

	}

}

// RandomExpiration 传入一个基础过期时间，返回增加 0~59 秒随机抖动后的时间
func RandomExpiration(baseTime time.Duration) time.Duration {
	// 使用 rand.Intn(60) 生成 0-59 的随机数，更加标准和易读
	jitter := time.Duration(rand.IntN(60)) * time.Second
	return baseTime + jitter
}
