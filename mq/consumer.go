package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"gggvrm/global"
	"log"

	"github.com/redis/go-redis/v9"
)

// LikeMessage 定义点赞消息体，必须和生产者发送的 JSON 字段一一对应
type LikeMessage struct {
	Action         string `json:"action"`
	ArticleID      int    `json:"article_id"`
	UserID         uint   `json:"user_id"`         // 谁点的赞
	AuthorUsername string `json:"author_username"` // 被点赞的文章作者
	Timestamp      int64  `json:"timestamp"`
}

// ArticleMessage 定义文章消息体，用于 Feed 流推送
type ArticleMessage struct {
	Action    string `json:"action"`     // 例如 "create_article"
	ArticleID uint   `json:"article_id"` // 文章ID
	UserID    uint   `json:"user_id"`    // 作者ID
	Timestamp int64  `json:"timestamp"`  // 发布时间戳，将作为 ZSet 的 Score
}

// FavoriteMessage 定义收藏消息体，用于收藏通知
type FavoriteMessage struct {
	Action         string `json:"action"`          // "favorite_article"
	ArticleID      int    `json:"article_id"`      // 文章ID
	UserID         uint   `json:"user_id"`         // 谁收藏的
	AuthorUsername string `json:"author_username"` // 文章作者用户名（生产者富化）
	Timestamp      int64  `json:"timestamp"`
}

// FollowMessage 定义关注消息体，用于关注通知
type FollowMessage struct {
	Action           string `json:"action"`            // "follow_user"
	FollowerID       uint   `json:"follower_id"`       // 谁关注的
	FollowerUsername string `json:"follower_username"` // 关注者用户名（生产者富化）
	FolloweeID       uint   `json:"followee_id"`       // 被关注的
	Timestamp        int64  `json:"timestamp"`
}

// StartConsumers 统一启动所有的 MQ 消费者
func StartConsumers() {
	ctx := context.Background()

	// ==========================================
	// 1. 启动 article_tasks 消费者（Feed 流推送）
	// ==========================================
	articleMsgs, err := global.RabbitMQChan.Consume(
		"article_tasks", "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("注册 article_tasks 消费者失败: %v", err)
	}

	go func() {
		for d := range articleMsgs {
			log.Printf("收到 article_tasks 消息: %s", d.Body)

			var msg ArticleMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("文章消息 JSON 解析失败: %v", err)
				d.Ack(false)
				continue
			}

			if msg.Action == "create_article" {
				// 1. 查询作者的所有粉丝 ID
				var followerIDs []uint
				err := global.Db.Table("user_follows").
					Where("followee_id = ?", msg.UserID).
					Pluck("follower_id", &followerIDs).Error
				if err != nil {
					log.Printf("【警告】查询粉丝列表失败: %v", err)
					d.Ack(false)
					continue
				}

				if len(followerIDs) == 0 {
					d.Ack(false)
					continue
				}

				// 2. 开启 Redis Pipeline，批量推送 Feed 流
				pipe := global.RedisDB.Pipeline()

				for _, followerID := range followerIDs {
					// 每个粉丝拥有一个专属的 ZSet 收件箱
					feedKey := fmt.Sprintf("feed:user:%d", followerID)

					// 将文章 ID 放入粉丝的收件箱，以时间戳作为排序依据
					pipe.ZAdd(ctx, feedKey, redis.Z{
						Score:  float64(msg.Timestamp),
						Member: msg.ArticleID,
					})

					// 控制信箱容量，防止大V粉丝的 Redis 内存撑爆
					// 只保留最近的 1000 条动态，清理掉旧的
					pipe.ZRemRangeByRank(ctx, feedKey, 0, -1001)
				}

				// 3. 一次性将所有命令打包发给 Redis
				if _, err := pipe.Exec(ctx); err != nil {
					log.Printf("【严重】Feed流推送 Redis Pipeline 执行失败: %v", err)
				} else {
					log.Printf("成功将文章 %d 推送给 %d 个粉丝", msg.ArticleID, len(followerIDs))
				}
			}

			d.Ack(false)
		}
	}()
	log.Println("RabbitMQ article_tasks 消费者已启动...")

	//global.Me.Publish是无差别广播，没改成单用户推送，暂时注释掉收藏通知，避免消息过多刷屏
	// ==========================================
	// 2. 启动 favorite_tasks 消费者（收藏通知）
	// ==========================================
	//favoriteMsgs, err := global.RabbitMQChan.Consume(
	//	"favorite_tasks", "", false, false, false, false, nil,
	//)
	//if err != nil {
	//	log.Fatalf("注册 favorite_tasks 消费者失败: %v", err)
	//}
	//go func() {
	//	for d := range favoriteMsgs {
	//		log.Printf("收到收藏消息: %s", d.Body)
	//
	//		var msg FavoriteMessage
	//		if err := json.Unmarshal(d.Body, &msg); err != nil {
	//			log.Printf("收藏消息 JSON 解析失败: %v", err)
	//			d.Ack(false)
	//			continue
	//		}
	//
	//		if msg.Action == "favorite_article" {
	//			// 通过 WebSocket 推送给文章作者（用户名已由生产者富化，零查库）
	//			if msg.AuthorUsername != "" {
	//				content := fmt.Sprintf("用户 %d 刚刚收藏了您的文章！", msg.UserID)
	//				global.Me.Publish(global.Message{
	//					Username: msg.AuthorUsername,
	//					Content:  content,
	//				})
	//			} else {
	//				log.Printf("消息中无作者用户名，跳过WebSocket通知 (文章ID: %d)", msg.ArticleID)
	//			}
	//		}
	//
	//		d.Ack(false)
	//	}
	//}()
	//log.Println("RabbitMQ favorite_tasks 消费者已启动...")

	//global.Me.Publish是无差别广播，没改成单用户推送，暂时注释掉通知，避免消息过多刷屏
	// ==========================================
	// 3. 启动 follow_tasks 消费者（关注通知）
	// ==========================================
	//followMsgs, err := global.RabbitMQChan.Consume(
	//	"follow_tasks", "", false, false, false, false, nil,
	//)
	//if err != nil {
	//	log.Fatalf("注册 follow_tasks 消费者失败: %v", err)
	//}
	//
	//go func() {
	//	for d := range followMsgs {
	//		log.Printf("收到关注消息: %s", d.Body)
	//
	//		var msg FollowMessage
	//		if err := json.Unmarshal(d.Body, &msg); err != nil {
	//			log.Printf("关注消息 JSON 解析失败: %v", err)
	//			d.Ack(false)
	//			continue
	//		}
	//
	//		if msg.Action == "follow_user" {
	//			// 查询被关注者的用户名（仅此一个查库，因为被关注者不在消息中）
	//			var followeeUsername string
	//			err := global.Db.Table("users").
	//				Select("username").
	//				Where("id = ?", msg.FolloweeID).
	//				Scan(&followeeUsername).Error
	//			if err != nil {
	//				log.Printf("【警告】查询被关注者用户名失败: %v", err)
	//				d.Ack(false)
	//				continue
	//			}
	//
	//			// 通过 WebSocket 推送给被关注者（关注者用户名已由生产者富化）
	//			if followeeUsername != "" {
	//				content := fmt.Sprintf("用户 %d 刚刚关注了您！", msg.FollowerID)
	//				if msg.FollowerUsername != "" {
	//					content = fmt.Sprintf("用户 %s 刚刚关注了您！", msg.FollowerUsername)
	//				}
	//				global.Me.Publish(global.Message{
	//					Username: followeeUsername,
	//					Content:  content,
	//				})
	//			}
	//		}
	//
	//		d.Ack(false)
	//	}
	//}()
	//log.Println("RabbitMQ follow_tasks 消费者已启动...")

	//global.Me.Publish是无差别广播，没改成单用户推送，暂时注释掉通知，避免消息过多刷屏
	// ==========================================
	// 4. 启动 like_tasks 消费者
	// ==========================================
	//likeMsgs, err := global.RabbitMQChan.Consume(
	//	"like_tasks", "", false, false, false, false, nil,
	//)
	//if err != nil {
	//	log.Fatalf("注册 like_tasks 消费者失败: %v", err)
	//}
	//
	//go func() {
	//	for d := range likeMsgs {
	//		log.Printf("收到点赞消息: %s", d.Body)
	//
	//		var msg LikeMessage
	//		// 1. 解析 JSON 数据
	//		if err := json.Unmarshal(d.Body, &msg); err != nil {
	//			log.Printf("点赞消息 JSON 解析失败: %v", err)
	//			// 关键点：即使解析失败也要 Ack，把脏数据扔掉，否则会卡死队列
	//			err := d.Ack(false)
	//			if err != nil {
	//				log.Printf("点赞消息 Ack 失败: %v", err)
	//				continue
	//			}
	//			continue
	//		}
	//
	//		if msg.AuthorUsername != "" {
	//			global.Me.Publish(global.Message{
	//				Username: msg.AuthorUsername,
	//				Content:  fmt.Sprintf("用户 %d 刚刚点赞了您的文章！", msg.UserID),
	//			})
	//		} else {
	//			log.Printf("消息中无作者用户名，跳过WebSocket通知 (文章ID: %d)", msg.ArticleID)
	//		}
	//
	//		// 4. 所有业务处理完毕后，安全地手动确认消息
	//		err := d.Ack(false)
	//		if err != nil {
	//			log.Printf("点赞消息 Ack 失败: %v", err)
	//			continue
	//		}
	//	}
	//}()
	//log.Println("RabbitMQ like_tasks 消费者已启动...")
}
