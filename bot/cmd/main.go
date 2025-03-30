package main

import (
	"context"
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

	ctx, finish := context.WithCancel(context.Background())

	bot := bot.NewBot(ctx, cfg)

	go bot.ListenToEvents(ctx)

	bot.GracefulShutdown()
	finish()
	bot.WaitСlosingProcesses()

}
