package global

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/go-redis/redis"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type MessageBroker interface {
	// 订阅：用户连入时，把他的专属信箱交给邮局登记
	Subscribe(subscriber chan Message)
	// 取消订阅：用户离开时，从邮局注销信箱
	Unsubscribe(subscriber chan Message)
	// 发布：不管是谁，只要调用这个方法，邮局就把消息塞进所有已登记的信箱
	Publish(msg Message)
	// 启动：开启邮局的分发流水线
	Start()
}

type LocalBroker struct {
	subscribers sync.Map
	broadcast   chan Message
}

func NewLocalBroker() *LocalBroker {
	return &LocalBroker{
		broadcast: make(chan Message, 256),
	}
}

// 登记信箱
func (l *LocalBroker) Subscribe(subscriber chan Message) {
	l.subscribers.Store(subscriber, true)
}

// 注销信箱
func (l *LocalBroker) Unsubscribe(subscriber chan Message) {
	_, ok := l.subscribers.LoadAndDelete(subscriber)
	if ok {
		close(subscriber)
	} else {
		log.Println("尝试注销一个不存在的订阅者")
	}
}

// 接收寄件请求
func (l *LocalBroker) Publish(msg Message) {
	l.broadcast <- msg
}

// 开启分发流水线（原先的 HandleMessages）
func (l *LocalBroker) Start() {
	for {
		msg := <-l.broadcast
		l.subscribers.Range(func(key, _ interface{}) bool {
			sendChan := key.(chan Message)
			select {
			case sendChan <- msg:
			default:
				// 信箱满了，说明这人卡死了，直接踢掉
				l.Unsubscribe(sendChan)
				log.Println("用户离开聊天室")
				return true
			}
			return true
		})
	}
}

type RedisBroker struct {
	rdb         *redis.Client
	channelName string
	subscribers sync.Map
}

func NewRedisBroker() *RedisBroker {
	return &RedisBroker{
		rdb:         RedisDB,
		channelName: "broadcast",
		subscribers: sync.Map{},
	}
}

// 登记信箱
func (r *RedisBroker) Subscribe(subscriber chan Message) {
	r.subscribers.Store(subscriber, true)
}

// 取消订阅：用户离开时，从邮局注销信箱
func (r *RedisBroker) Unsubscribe(subscriber chan Message) {
	_, ok := r.subscribers.LoadAndDelete(subscriber)
	if ok {
		close(subscriber)
	} else {
		log.Println("尝试注销一个不存在的订阅者")
	}
}

// 发布：不管是谁，只要调用这个方法，邮局就把消息塞进所有已登记的信箱
func (r *RedisBroker) Publish(msg Message) {
	msgbyte, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return
	}
	err = r.rdb.Publish(r.channelName, msgbyte).Err()
	if err != nil {
		log.Println("向 Redis 发布消息失败:", err)
	}
}

// 启动：开启邮局的分发流水线
func (r *RedisBroker) Start() {
	pubsub := r.rdb.Subscribe(r.channelName)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for redisMsg := range ch {
		var msg Message
		err := json.Unmarshal([]byte(redisMsg.Payload), &msg)
		if err != nil {
			log.Println(err)
			continue
		}
		r.subscribers.Range(func(key, _ interface{}) bool {
			sendChan := key.(chan Message)
			select {
			case sendChan <- msg:
			default:
				// 信箱满了，说明这人卡死了，直接踢掉
				r.Unsubscribe(sendChan)
				log.Println("用户离开聊天室")
				return true
			}
			return true
		})
	}
}
