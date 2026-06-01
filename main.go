package main

import (
	"context"
	"errors"
	"gggvrm/config"
	"gggvrm/controllers"
	"gggvrm/global"
	"gggvrm/mq"
	"gggvrm/rag_grpc"
	"gggvrm/router"
	"gggvrm/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	config.InitConfig()

	r := router.SetupRouter()

	Port := config.Appconf.App.Port
	if Port == "" {
		Port = ":8080"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go utils.SyncSql(ctx) //同步like数据到数据库

	mq.StartConsumers()

	if config.Appconf.Database.SubSwitch {
		global.Me = global.NewRedisBroker()
	} else {
		global.Me = global.NewLocalBroker()
	}

	go controllers.HandleMessages()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接 Python gRPC 服务: %v", err)
	}
	defer conn.Close()
	client := rag_grpc.NewRagServiceClient(conn)
	global.Rag_grpc_client = client

	srv := &http.Server{
		Addr:    Port,
		Handler: r,
	}

	go func() {
		log.Printf("服务器正在%s端口运行 \n", Port)
		// ErrServerClosed 是调用 Shutdown 后的正常返回，需要过滤掉以免报 err
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt) //监听系统中断信号
	<-quit
	log.Println("服务器正在关闭...")
	cancel()
	log.Println("已通知后台任务停止，等待 HTTP 请求处理完成...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("服务器强行关闭或超时异常: ", err)
	}

	log.Println("服务器已成功优雅退出")
}
