package controllers

import (
	"gggvrm/global"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	pingPeriod = 30 * time.Second // 多久发一次 Ping
	pongWait   = 60 * time.Second // 绝对超时时间（超过这个时间没收到响应就踢人）
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

	var send = make(chan global.Message, 256) // 专属的发送通道（信箱）

	// 将新连接的客户端添加到 全局订阅列表中，这样当有消息广播时就能收到
	global.Me.Subscribe(send)

	log.Println("有新用户加入聊天室")

	go writePump(ws, send) // 启动专属写协程，负责给这个客户端发消息

	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		// 只要收到客户端的 Pong 回复，就重置超时倒计时（续命）
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 不断监听该客户端发来的消息
	for {
		var msg global.Message
		// 读取客户端发来的 JSON 数据并解析到 msg 结构体
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("读取消息错误或用户离开:", err)
			// 如果读取出错（通常是用户断开连接），从 map 中移除该连接并结束循环
			global.Me.Unsubscribe(send)
			break
		}

		msg.Username = val.(string)

		// 把收到的消息塞进广播通道
		global.Me.Publish(msg)
	}
}

// 监听广播通道，把消息推送给所有人
func HandleMessages() {
	global.Me.Start()
}

// 专属写协程：只负责给这一个特定的客户端发消息
func writePump(client *websocket.Conn, send chan global.Message) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Close()
	}()

	for {
		select {
		case msg, ok := <-send: // 1. 有人往他的信箱里塞了聊天消息
			if !ok {
				// 通道被关闭，说明要断开连接了
				client.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// 发送聊天消息
			err := client.WriteJSON(msg)
			if err != nil {
				return
			}

		case <-ticker.C: // 2. 定时器响了，该发心跳 Ping 了
			// 每次发 Ping 之前，可以设置一个写入超时时间
			client.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // 发送 Ping 失败，直接退出并关闭连接
			}
		}
	}
}
