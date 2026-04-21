package mq

import (
	"gggvrm/global"
	"log"
)

// StartConsumers 消费者
func StartConsumers() {
	msgs, err := global.RabbitMQChan.Consume(
		"article_tasks", // 队列名称
		"",              // consumer
		false,           // auto-ack (设为 false，确保手动确认，防止消息丢失)
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			log.Printf("收到 RabbitMQ 消息: %s", d.Body)

			// 在这里执行耗时任务，例如：处理数据库持久化、发送邮件通知等
			// ... 业务逻辑 ...

			// 任务处理完成后，手动确认消息已消费
			d.Ack(false)
		}
	}()
	log.Println("RabbitMQ 消费者已启动，正在监听消息...")
}
