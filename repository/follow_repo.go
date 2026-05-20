package repository

import (
	"context"
	"gggvrm/models"

	"gorm.io/gorm"
)

// FollowRepository 关注数据访问接口
type FollowRepository interface {
	Follow(ctx context.Context, followerID, followeeID uint) error                                     // 关注用户
	Unfollow(ctx context.Context, followerID, followeeID uint) error                                   // 取消关注
	IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error)                        // 查询是否已关注
	GetFollowingCount(ctx context.Context, userID uint) (int64, error)                                 // 获取用户的关注数
	GetFollowersCount(ctx context.Context, userID uint) (int64, error)                                 // 获取用户的粉丝数
	GetFollowing(ctx context.Context, userID uint, offset, pageSize int) ([]models.User, int64, error) // 分页获取用户的关注列表
	GetFollowers(ctx context.Context, userID uint, offset, pageSize int) ([]models.User, int64, error) // 分页获取用户的粉丝列表
}

type followRepoImpl struct {
	db *gorm.DB
}

// NewFollowRepository 创建关注仓库实例
func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepoImpl{db: db}
}

// Follow 向 user_follows 表插入一条关注记录
func (r *followRepoImpl) Follow(ctx context.Context, followerID, followeeID uint) error {
	follow := models.UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	return r.db.WithContext(ctx).Create(&follow).Error
}

// Unfollow 从 user_follows 表删除一条关注记录
func (r *followRepoImpl) Unfollow(ctx context.Context, followerID, followeeID uint) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&models.UserFollow{}).Error
}

// IsFollowing 检查 follower 是否已关注 followee
func (r *followRepoImpl) IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowingCount 统计用户关注的人数
func (r *followRepoImpl) GetFollowingCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetFollowersCount 统计用户的粉丝数
func (r *followRepoImpl) GetFollowersCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserFollow{}).
		Where("followee_id = ?", userID).
		Count(&count).Error
	return count, err
}

// GetFollowing 通过中间表关联查询用户关注的人列表，支持分页
func (r *followRepoImpl) GetFollowing(ctx context.Context, userID uint, offset, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).
		Joins("JOIN user_follows ON user_follows.followee_id = users.id").
		Where("user_follows.follower_id = ?", userID)

	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 初始化为空切片，防止返回 null
	users = make([]models.User, 0)

	if total > 0 && int64(offset) < total {
		if err := query.
			Order("user_follows.created_at DESC").
			Limit(pageSize).Offset(offset).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
	}

	return users, total, nil
}

// GetFollowers 通过中间表关联查询用户的粉丝列表，支持分页
func (r *followRepoImpl) GetFollowers(ctx context.Context, userID uint, offset, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).
		Joins("JOIN user_follows ON user_follows.follower_id = users.id").
		Where("user_follows.followee_id = ?", userID)

	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 初始化为空切片，防止返回 null
	users = make([]models.User, 0)

	if total > 0 && int64(offset) < total {
		if err := query.
			Order("user_follows.created_at DESC").
			Limit(pageSize).Offset(offset).
			Find(&users).Error; err != nil {
			return nil, 0, err
		}
	}

	return users, total, nil
}
