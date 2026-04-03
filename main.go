package main

import (
	"gggvrm/config"
	"gggvrm/router"
)

func main() {
	config.InitConfig()

	r := router.SetupRouter()

	Port := config.Appconf.App.Port
	if Port == "" {
		Port = ":8080"
	}

	err := r.Run(Port)
	if err != nil {
		return 
	}
}
