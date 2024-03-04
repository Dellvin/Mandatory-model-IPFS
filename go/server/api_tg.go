package main

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"server/storage"
)

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	l := ctx.Value("logger").(*zap.Logger)
	switch update.Message.Text {
	case "/start":
		helloMessage(ctx, b, update, l)
	default:
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   update.Message.Text,
		})
		if err != nil {
			l.Error("failed to send message", zap.Error(err))
		}
	}
}

func helloMessage(ctx context.Context, b *bot.Bot, update *models.Update, l *zap.Logger) {
	user := ctx.Value("user").(*storage.User)
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Hello %s\nYour department: %d\nYour level: %d\nIf there is a mistake, please contact your system adminstrator.", user.TgName, user.Department, user.Level),
	})
	if err != nil {
		l.Error("failed to send message", zap.Error(err))
	}
}
