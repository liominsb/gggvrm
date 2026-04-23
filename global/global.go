package global // Package global 全局的

import (
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

var (
	Db           *gorm.DB //database 数据库
	RedisDB      *redis.Client
	RabbitMQConn *amqp.Connection
	RabbitMQChan *amqp.Channel
	Rekg         bool
	Me           MessageBroker
)
