package controllers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

// 全局变量维护聊天室状态
var (
	// 用于保存所有连接的客户端（使用 sync.Map 保证并发安全）
	clients sync.Map
	// 广播通道：当有用户发送消息时，把消息丢进这个通道
	broadcast = make(chan Message)
)

// 定义 WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源（生产环境需限制）
		return true
	},
}

// 处理客户端发起的 WebSocket 连接请求
func HandleConnections(ctx *gin.Context) {
	val, ok := ctx.Get("ID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录或身份已过期"})
		return
	}

	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println("协议升级失败:", err)
		return
	}
	defer ws.Close()

	clients.Store(ws, true) // 将新连接的客户端添加到 clients 中
	log.Println("有新用户加入聊天室")

	// 不断监听该客户端发来的消息
	for {
		var msg Message
		// 读取客户端发来的 JSON 数据并解析到 msg 结构体
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("读取消息错误或用户离开:", err)
			// 如果读取出错（通常是用户断开连接），从 map 中移除该连接并结束循环
			clients.Delete(ws)
			break
		}

		msg.Username = val.(string)

		// 把收到的消息塞进广播通道
		broadcast <- msg
	}
}

// 监听广播通道，把消息推送给所有人
func HandleMessages() {
	for {
		// 从通道中取出消息（如果通道里没消息，这里会阻塞等待）
		msg := <-broadcast

		// 遍历所有在线的客户端，把消息发给他们
		clients.Range(func(key, _ interface{}) bool {
			client := key.(*websocket.Conn)

			// 将消息编码为 JSON 并发送给客户端
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("发送消息失败，清理该连接:", err)
				client.Close()
				clients.Delete(client)
			}
			return true // 返回 true 继续遍历下一个
		})
	}
}
