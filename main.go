package main

import (
	"gggvrm/config"
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

	go utils.SyncSql() //同步like数据到数据库

	err := r.Run(Port)
	if err != nil {
		return
	}
}
