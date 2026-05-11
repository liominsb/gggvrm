package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gggvrm/models"
	"gggvrm/repository"
	"gggvrm/utils"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type AuthService interface {
	Register(ctx context.Context, Username string, Password string) (string, error)
	Login(ctx context.Context, Username string, Password string) (string, error)
	GetMyUser(ctx context.Context, userID uint) (*models.User, error)
	GetUserProfileById(ctx context.Context, targetUserID uint) (*models.User, error)
	ChangePassword(ctx context.Context, userID uint, oldPassword string, newPassword string) error
}

type authServiceImpl struct {
	authRepo    repository.AuthRepository
	redisClient *redis.Client
	sfGroup     singleflight.Group
}

func NewAuthService(authRepo repository.AuthRepository, redisClient *redis.Client) AuthService {
	return &authServiceImpl{authRepo: authRepo, redisClient: redisClient}
}

func (s *authServiceImpl) Register(ctx context.Context, Username string, Password string) (string, error) {
	var user models.User
	hashedPwd, err := utils.HashPassword(Password)
	if err != nil {
		return "", err
	}

	user.Username = Username
	user.Password = hashedPwd

	if err := s.authRepo.Register(ctx, &user); err != nil {
		return "", err
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authServiceImpl) Login(ctx context.Context, Username string, Password string) (string, error) {
	var user models.User
	if err := s.authRepo.GetUserByUsername(ctx, &user, Username); err != nil {
		return "", err
	}
	if !utils.CheckPassword(Password, user.Password) {
		return "", errors.New("密码错误")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", errors.New("生成 token 失败")
	}
	return token, nil
}

func (s *authServiceImpl) GetMyUser(ctx context.Context, userID uint) (*models.User, error) {
	cacheKey := fmt.Sprintf("USER:%d", userID)

	// 先尝试从Redis缓存获取
	data, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中
		var user models.User
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
		user.Password = ""
		return &user, nil
	}

	// 缓存未命中，检查是否是redis.Nil错误
	if !errors.Is(err, redis.Nil) {
		// 其他Redis错误
		return nil, err
	}

	// 从数据库获取用户
	var user models.User
	if err := s.authRepo.GetUserByID(ctx, &user, userID); err != nil {
		return nil, err
	}

	// 将用户信息存入缓存
	if err := utils.Setcache(ctx, cacheKey, user); err != nil {
		// 缓存设置失败不影响主流程，只记录日志
		fmt.Printf("设置缓存失败 UserID: %d, err: %v\n", userID, err)
	}

	// 返回用户信息（不包含密码）
	user.Password = ""
	return &user, nil
}

func (s *authServiceImpl) GetUserProfileById(ctx context.Context, targetUserID uint) (*models.User, error) {
	cacheKey := fmt.Sprintf("USER:%d", targetUserID)

	// 先尝试从Redis缓存获取
	data, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中
		var user models.User
		if err := json.Unmarshal([]byte(data), &user); err != nil {
			return nil, err
		}
		user.Password = ""
		return &user, nil
	}

	// 缓存未命中，检查是否是redis.Nil错误
	if !errors.Is(err, redis.Nil) {
		// 其他Redis错误
		return nil, err
	}

	// 从数据库获取用户
	var user models.User
	if err := s.authRepo.GetUserByID(ctx, &user, targetUserID); err != nil {
		return nil, err
	}

	// 将用户信息存入缓存
	if err := utils.Setcache(ctx, cacheKey, user); err != nil {
		// 缓存设置失败不影响主流程，只记录日志
		fmt.Printf("设置缓存失败 UserID: %d, err: %v\n", targetUserID, err)
	}

	// 返回用户信息（不包含密码）
	user.Password = ""
	return &user, nil
}

func (s *authServiceImpl) ChangePassword(ctx context.Context, userID uint, oldPassword string, newPassword string) error {
	// 获取用户信息
	var user models.User
	if err := s.authRepo.GetUserByID(ctx, &user, userID); err != nil {
		return err
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("旧密码不正确")
	}

	// 检查新密码是否与旧密码相同
	if oldPassword == newPassword {
		return errors.New("新密码不能与旧密码相同")
	}

	// 加密新密码
	hashedPwd, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// 更新密码
	if err := s.authRepo.UpdatePassword(ctx, userID, hashedPwd); err != nil {
		return err
	}

	// 删除缓存
	cacheKey := fmt.Sprintf("USER:%d", userID)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		fmt.Printf("删除缓存失败 UserID: %d, err: %v\n", userID, err)
	}

	return nil
}
