package pkg

import (
	"fmt"
	"github.com/go-telegram/bot/models"
)

type TgData struct {
	Username   string
	Text       string
	IsCallback bool
	ChatID     int64
}

func GetTgData(update *models.Update) (TgData, error) {
	tg := TgData{}
	if update == nil {
		return tg, fmt.Errorf("message is nil")
	} else if update.Message != nil {
		if update.Message.From != nil {
			return TgData{
				Username: update.Message.From.Username,
				Text:     update.Message.Text,
				ChatID:   update.Message.Chat.ID,
			}, nil
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Message != nil && update.CallbackQuery.Message.Message.From != nil {
			return TgData{
				Username:   update.CallbackQuery.Message.Message.Chat.Username,
				Text:       update.CallbackQuery.Data,
				ChatID:     update.CallbackQuery.Message.Message.Chat.ID,
				IsCallback: true,
			}, nil
		}
	}

	return tg, fmt.Errorf("failed to get data")
}
