package utils // Package utils 实用性

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/config"
	"gggvrm/global"
	"gggvrm/models"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	return string(hash), err
}

func GenerateJWT(id uint) (string, error) { //生成JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":  id,
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	})
	signedToken, err := token.SignedString([]byte(config.Appconf.JWT.Key))
	return "Bearer " + signedToken, err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ParseJWT 解析JWT
func ParseJWT(tokenString string) (uint, error) { //
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Appconf.JWT.Key), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["ID"].(float64)
		if !ok {
			return 0, errors.New("invalid token")
		}
		return uint(id), nil
	}
	return 0, errors.New("invalid token")
}

// Setcache 设置缓存
func Setcache(key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		return err
	}
	a := time.Duration(rand.IntN(5) + 10)
	if err := global.RedisDB.Set(key, valueJSON, a*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

func SyncSql() {
	for {
		time.Sleep(1 * time.Minute)
		var cursor uint64 // 初始游标为 0
		var cursor1 uint64
		var keys []string
		var err error
		for {
			keys, cursor, err = global.RedisDB.Scan(cursor, "article:*:likes", 100).Result()
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

				likesStr, err := global.RedisDB.Get(key).Result()
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
			keys, cursor1, err = global.RedisDB.Scan(cursor1, "article:*:views", 100).Result()
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

				viewsStr, err := global.RedisDB.Get(key).Result()
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
