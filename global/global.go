package global // Package global 全局的

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	Db      *gorm.DB //database 数据库
	RedisDB *redis.Client
)
