package global // Package global 全局的

import (
	"gggvrm/rag_grpc"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	Db              *gorm.DB //database 数据库
	RedisDB         *redis.Client
	RabbitMQConn    *amqp.Connection
	RabbitMQChan    *amqp.Channel
	Me              MessageBroker //RedisBroker和LocalBroker的通用接口
	Rag_grpc_client rag_grpc.RagServiceClient
)
