package main

import (
	"context"
	"gggvrm/config"
	"gggvrm/controllers"
	"gggvrm/global"
	"gggvrm/router"
	"gggvrm/utils"
)

func main() {
	config.InitConfig()

	r := router.SetupRouter()

	Port := config.Appconf.App.Port
	if Port == "" {
		Port = ":8080"
	}

	var ctx context.Context = context.Background()

	go utils.SyncSql(ctx) //同步like数据到数据库

	go controllers.HandleMessages()

	if config.Appconf.Database.SubSwitch {
		global.Me = global.NewRedisBroker()
	} else {
		global.Me = global.NewLocalBroker()
	}

	err := r.Run(Port)
	if err != nil {
		return
	}
}
