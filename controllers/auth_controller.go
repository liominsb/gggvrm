package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/global"
	"gggvrm/models"
	"gggvrm/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// Register 注册
func Register(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBind(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPwd, err := utils.HashPassword(user.Password)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = hashedPwd

	token, err := utils.GenerateJWT(user.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := global.Db.AutoMigrate(&models.User{}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to migrate database"})
		return
	}

	if err := global.Db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create database"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// Login 登录
func Login(ctx *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBind(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User

	if err := global.Db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(user.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// Getmyuser 查当前登录的用户信息
func Getmyuser(ctx *gin.Context) {
	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	var user models.User

	data, err := global.RedisDB.Get(fmt.Sprintf("USER:%d", id.(uint))).Result()

	if errors.Is(err, redis.Nil) {
		if err := global.Db.Where("id = ?", id).First(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		utils.Setcache(ctx, fmt.Sprintf("USER:%d", id.(uint)), user)
		user.Password = ""
		ctx.JSON(http.StatusOK, gin.H{"user": user})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Password = ""
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

// GetUserProfileById 通过动态 ID 获取其他用户的信息
func GetUserProfileById(ctx *gin.Context) {
	// 1. 从 URL 参数中获取 ID (例如请求 /api/v1/user/5，这里拿到的就是 "5")
	idStr := ctx.Param("id")

	// 把字符串转换成整数
	targetUserID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID格式"})
		return
	}

	var user models.User
	// 为了和当前用户自己的缓存区分开，可以加个前缀或直接用一样的
	cacheKey := fmt.Sprintf("USER:%d", targetUserID)

	data, err := global.RedisDB.Get(cacheKey).Result()

	// 2. 缓存未命中，去数据库查
	if errors.Is(err, redis.Nil) {
		if err := global.Db.Where("id = ?", targetUserID).First(&user).Error; err != nil {
			// 如果查不到，返回 404
			ctx.JSON(http.StatusNotFound, gin.H{"error": "找不到该用户"})
			return
		}

		// 绝对不能把密码暴露出去
		user.Password = ""

		utils.Setcache(ctx, cacheKey, user)
		ctx.JSON(http.StatusOK, gin.H{"user": user})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 缓存命中，解析 JSON
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "数据解析异常"})
		return
	}

	// 再次确保密码为空
	user.Password = ""
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func Changepassword(ctx *gin.Context) {
	id, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ID"})
		return
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := global.Db.Where("id = ?", id).First(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !utils.CheckPassword(input.OldPassword, user.Password) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Old password is incorrect"})
		return
	}

	if input.OldPassword == input.NewPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "New password cannot be the same as the old password"})
		return
	}

	hashedPwd, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = hashedPwd
	if err := global.Db.Model(&user).Update("password", user.Password).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := global.RedisDB.Del(fmt.Sprintf("USER:%d", id.(uint))).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}
