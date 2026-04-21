package config

import (
	"gggvrm/global"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func initRabbitMQ() {
	var err error
	global.RabbitMQConn, err = amqp.Dial(Appconf.RabbitMQ.Url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	global.RabbitMQChan, err = global.RabbitMQConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	// 声明一个常用的队列（例如：处理文章系统消息）
	_, err = global.RabbitMQChan.QueueDeclare(
		"article_tasks", // 队列名称
		true,            // durable (持久化)
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	_, err = global.RabbitMQChan.QueueDeclare(
		"like_tasks", // 队列名称
		true,         // durable (持久化)
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	log.Println("RabbitMQ 初始化成功！")
}
