package telegramapi

import (
	"dolphin-bot/internal/telegram-api/handlers"
	"time"

	tele "gopkg.in/telebot.v3"
)

const poolerTimeout = 10 * time.Second

type Bot struct {
	telebot *tele.Bot
}

func NewBot(token string) *Bot {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: poolerTimeout},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		panic(err)
	}

	setupHandlers(b)

	return &Bot{
		telebot: b,
	}
}

func (b *Bot) Start() {
	b.telebot.Start()
}

func (b *Bot) Stop() {
	b.telebot.Stop()
}

func setupHandlers(telebot *tele.Bot) {
	telebot.Handle("/start", handlers.HandleStart)
	telebot.Handle(tele.OnVoice, handlers.HandleVoiceMsg)
}
