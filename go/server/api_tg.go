package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"server/ipfs"
	"server/pkg"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	"server/storage"
)

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	l := ctx.Value("logger").(*zap.Logger)
	data := ctx.Value("data").(*pkg.TgData)
	user := ctx.Value("user").(*storage.User)
	switch {
	case data.Text == "/start":
		helloMessage(ctx, b, data, l)
	case update.Message.Document != nil:
		f, err := b.GetFile(ctx, &bot.GetFileParams{FileID: update.Message.Document.FileID})
		if err != nil {
			l.Error("failed to GetFile", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to GetFile")})
			return
		}

		resp, err := http.Get(b.FileDownloadLink(f))
		if err != nil {
			l.Error("failed to Get", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to Get")})
			return
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Error("failed to ReadAll", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to ReadAll")})
			return
		}
		if err = upload(raw, user, storage.File{Name: update.Message.Document.FileName, MimeType: update.Message.Document.MimeType, Type: "Document"}); err != nil {
			l.Error("failed to upload", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to upload")})
			return
		}
	default:
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "Список файлов", CallbackData: "list"}}, {{Text: "Скачать файл", CallbackData: "download"}},
			},
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      data.ChatID,
			Text:        "Выберете действие",
			ReplyMarkup: kb,
		})
	}
}

func upload(file []byte, user *storage.User, fileMeta storage.File) error {
	base64Cipher, err := enc(storage.User{
		ID:         user.ID,
		TgName:     "",
		PK:         user.PK,
		Department: user.Department,
		Level:      user.Level,
	}, string(file))
	if err != nil {
		return fmt.Errorf("failed to enc: %w", err)
	}

	link, err := ipfs.Upload("", []byte(base64Cipher))
	if err != nil {
		return fmt.Errorf("failed to Upload: %s", err.Error())
	}

	if err = storage.AddFile(db.DB, storage.File{
		Name:    fileMeta.Name,
		IpfsKey: link,
		UserID:  user.ID,
	}); err != nil {
		return fmt.Errorf("failed to AddFile: %s", err.Error())
	}

	return nil
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
