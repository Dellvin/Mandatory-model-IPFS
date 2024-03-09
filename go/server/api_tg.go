package main

import (
	"context"
	"fmt"
	"server/pkg"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"server/storage"
)

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	l := ctx.Value("logger").(*zap.Logger)
	data := ctx.Value("data").(*pkg.TgData)
	switch data.Text {
	case "/start":
		helloMessage(ctx, b, data, l)
	default:
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Список файлов", CallbackData: "list"},
					{Text: "Загрузить файл", CallbackData: "upload"},
				}, {
					{Text: "Скачать файл", CallbackData: "download"},
				},
			},
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      data.ChatID,
			Text:        "Выберете действие",
			ReplyMarkup: kb,
		})
	}
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	//data := ctx.Value("data").(*TgData)
	//user:=ctx.Value("user").(*storage.User)

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	switch update.CallbackQuery.Data {
	case "upload":

	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "You selected the button: " + update.CallbackQuery.Data,
	})
}

func helloMessage(ctx context.Context, b *bot.Bot, data *pkg.TgData, l *zap.Logger) {
	user := ctx.Value("user").(*storage.User)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: data.ChatID,
		Text:   fmt.Sprintf("Hello %s\nYour department: %d\nYour level: %d\nIf there is a mistake, please contact your system adminstrator.", user.TgName, user.Department, user.Level),
	})
	if err != nil {
		l.Error("failed to send message", zap.Error(err))
	}
}
