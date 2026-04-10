package utils // Package utils 实用性

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/config"
	"gggvrm/global"
	"gggvrm/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// 解析JWT
func ParseJWT(tokenString string) (uint, error) { //
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Appconf.JWT.Key), nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, ok := claims["ID"].(float64)
		if !ok {
			return 0, errors.New("Invalid token")
		}
		return uint(id), nil
	}
	return 0, errors.New("Invalid token")
}

// 设置缓存
func Setcache(ctx *gin.Context, key string, value interface{}) error {
	valueJSON, err := json.Marshal(value)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	if err := global.RedisDB.Set(key, valueJSON, 10*time.Minute).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return err
	}

	return nil
}

func SyncSql() {
	for {
		time.Sleep(1 * time.Minute)
		keys, err := global.RedisDB.Keys("article:*:likes").Result()
		if err != nil {
			fmt.Println("获取 Redis Keys 失败:", err)
			return
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

			if err := global.Db.Model(&models.Article{}).Update("likes", likes).Where("id = ?", articleID).Error; err != nil {
				fmt.Printf("更新文章 %d 点赞数失败: %v\n", articleID, err)
			}
		}
		fmt.Println("已同步点赞数到数据库")

		keys, err = global.RedisDB.Keys("article:*:views").Result()
		if err != nil {
			fmt.Println("获取 Redis Keys 失败:", err)
			return
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

			if err := global.Db.Model(&models.Article{}).Update("views", views).Where("id = ?", articleID).Error; err != nil {
				fmt.Printf("更新文章 %d 浏览数失败: %v\n", articleID, err)
			}
		}
		fmt.Println("已同步浏览数到数据库")
	}
}
