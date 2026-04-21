package mq

import (
	"context"
	"gggvrm/global"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishMessage 生产者，发送消息到指定队列
func PublishMessage(queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := global.RabbitMQChan.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // 消息持久化
			Body:         body,
		})
	return err
}
