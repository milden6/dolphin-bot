package main

import (
	"dolphin-bot/internal/config"
	telegramapi "dolphin-bot/internal/telegram-api"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := config.GetConfig()

	bot := telegramapi.NewBot(cfg.Token)

	go func() {
		bot.Start()
	}()

	slog.Info("Bot started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Service started")
	<-quit

	slog.Info("Stopping bot")
	bot.Stop()

	slog.Info("Bot stopped")
	slog.Info("Service stopped")
}
