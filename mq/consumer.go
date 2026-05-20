package mq

import (
	"encoding/json"
	"fmt"
	"gggvrm/global"
	"log"
)

// LikeMessage 定义点赞消息体，必须和生产者发送的 JSON 字段一一对应
type LikeMessage struct {
	Action         string `json:"action"`
	ArticleID      int    `json:"article_id"`
	UserID         uint   `json:"user_id"`         // 谁点的赞
	AuthorUsername string `json:"author_username"` // 【新增】被点赞的文章作者
	Timestamp      int64  `json:"timestamp"`
}

type ArticleMessage struct {
	Action    string `json:"action"`     // 例如 "create_article"
	ArticleID uint   `json:"article_id"` // 文章ID
	UserID    uint   `json:"user_id"`    // 作者ID
	Timestamp int64  `json:"timestamp"`  // 发布时间戳，将作为 ZSet 的 Score
}

// StartConsumers 统一启动所有的 MQ 消费者
func StartConsumers() {
	// ==========================================
	// 1. 启动 article_tasks 消费者
	// ==========================================
	//articleMsgs, err := global.RabbitMQChan.Consume(
	//	"article_tasks", "", false, false, false, false, nil,
	//)
	//if err != nil {
	//	log.Fatalf("注册 article_tasks 消费者失败: %v", err)
	//}

	//// ... 在 StartConsumers() 的 article_tasks 协程中：
	//go func() {
	//	for d := range articleMsgs {
	//		log.Printf("收到 article_tasks 消息: %s", d.Body)
	//
	//		var msg ArticleMessage
	//		if err := json.Unmarshal(d.Body, &msg); err != nil {
	//			log.Printf("文章消息 JSON 解析失败: %v", err)
	//			d.Ack(false)
	//			continue
	//		}
	//
	//		if msg.Action == "create_article" {
	//			// 1. 查询作者的所有粉丝 (这里假设你以后会建一个 user_follows 表)
	//			// 为了代码能跑，这里写一段示意 SQL。你需要根据实际表结构调整。
	//			var followerIDs []uint
	//			// err := global.Db.Table("user_follows").Where("followee_id = ?", msg.UserID).Pluck("follower_id", &followerIDs).Error
	//
	//			// 假设查到了粉丝：
	//			// if err == nil && len(followerIDs) > 0 {
	//
	//			// 2. 开启 Redis Pipeline，批量推送 Feed 流
	//			ctx := context.Background()
	//			pipe := global.RedisDB.Pipeline()
	//
	//			for _, followerID := range followerIDs {
	//				// 每个粉丝拥有一个专属的 ZSet 收件箱
	//				feedKey := fmt.Sprintf("feed:user:%d", followerID)
	//
	//				// 将文章ID放入粉丝的收件箱，以时间戳作为排序依据
	//				pipe.ZAdd(ctx, feedKey, redis.Z{
	//					Score:  float64(msg.Timestamp),
	//					Member: msg.ArticleID,
	//				})
	//
	//				// 【高端操作】：控制信箱容量，防止大V粉丝的 Redis 内存撑爆
	//				// 只保留最近的 1000 条动态，清理掉旧的
	//				pipe.ZRemRangeByRank(ctx, feedKey, 0, -1001)
	//			}
	//
	//			// 3. 一次性将所有命令打包发给 Redis
	//			_, err := pipe.Exec(ctx)
	//			if err != nil {
	//				log.Printf("【严重】Feed流推送 Redis Pipeline 执行失败: %v", err)
	//				// 视业务容忍度而定，这里可以选择不 Ack 让它重试，或者记录死信队列
	//			} else {
	//				log.Printf("成功将文章 %d 推送给 %d 个粉丝", msg.ArticleID, len(followerIDs))
	//			}
	//		}
	//
	//		d.Ack(false)
	//	}
	//}()
	//log.Println("RabbitMQ article_tasks 消费者已启动...")

	// ==========================================
	// 2. 启动 like_tasks 消费者
	// ==========================================
	likeMsgs, err := global.RabbitMQChan.Consume(
		"like_tasks", "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("注册 like_tasks 消费者失败: %v", err)
	}

	go func() {
		for d := range likeMsgs {
			log.Printf("收到点赞消息: %s", d.Body)

			var msg LikeMessage
			// 1. 解析 JSON 数据
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("点赞消息 JSON 解析失败: %v", err)
				// 关键点：即使解析失败也要 Ack，把脏数据扔掉，否则会卡死队列
				err := d.Ack(false)
				if err != nil {
					log.Printf("点赞消息 Ack 失败: %v", err)
					continue
				}
				continue
			}

			if msg.AuthorUsername != "" {
				global.Me.Publish(global.Message{
					Username: msg.AuthorUsername,
					Content:  fmt.Sprintf("用户 %d 刚刚点赞了您的文章！", msg.UserID), // 这里你可以改成查到点赞人的用户名再发，体验更好
				})
			} else {
				log.Printf("消息中无作者用户名，跳过WebSocket通知 (文章ID: %d)", msg.ArticleID)
			}

			// 4. 所有业务处理完毕后，安全地手动确认消息
			err := d.Ack(false)
			if err != nil {
				log.Printf("点赞消息 Ack 失败: %v", err)
				continue
			}
		}
	}()
	log.Println("RabbitMQ like_tasks 消费者已启动...")
}
