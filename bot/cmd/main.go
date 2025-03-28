package main

import (
	"log"
	"mmbot/internal/bot"
	"mmbot/internal/config"
	"mmbot/pkg/logger"
)

func main() {

	const methodPointer = "main.main"

	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("init logger failed: [%s]\n", err.Error())
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatal("get config failed", methodPointer, "text error", err.Error())
	}

	bot := bot.NewBot(cfg)

	bot.SetupGracefulShutdown()

	bot.ListenToEvents()

}
