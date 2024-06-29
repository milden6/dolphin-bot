package handlers

import (
	tele "gopkg.in/telebot.v3"
)

func HandleStart(ctx tele.Context) error {
	img := &tele.Photo{
		File:    tele.FromDisk("static/images/start_dolphin.png"),
		Caption: "Привет! 🤗 \n\nЯ веселый Dolphin! Отправь мне голосовое сообщение, а я попробую разобрать, что там 😎"}

	return ctx.Send(img, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
