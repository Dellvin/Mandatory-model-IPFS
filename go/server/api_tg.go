package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"server/ipfs"
	"server/pkg"
	"strconv"

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
	case update.Message != nil && update.Message.Document != nil:
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
				{{Text: "Список файлов", CallbackData: "button list"}}, {{Text: "Скачать файл", CallbackData: "button download"}},
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

func callbackMenuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := ctx.Value("data").(*pkg.TgData)
	user := ctx.Value("user").(*storage.User)
	l := ctx.Value("logger").(*zap.Logger)

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	switch update.CallbackQuery.Data {
	case "button list":
		files, err := storage.GetAccessedFiles(db.DB, *user)
		if err != nil {
			l.Error("failed to ReadAll", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to ReadAll")})
			return
		}
		var output = "List of available files"
		for _, f := range files {
			output = fmt.Sprintf("%s\nName: <b>%s</b>, Type: <b>%s</b>", output, f.Name, f.MimeType)
		}
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: output, ParseMode: models.ParseModeHTML})
	case "button download":
		files, err := storage.GetAccessedFiles(db.DB, *user)
		if err != nil {
			l.Error("failed to ReadAll", zap.Error(err))
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to ReadAll")})
			return
		}
		buttons := [][]models.InlineKeyboardButton{}
		for _, file := range files {
			buttons = append(buttons, []models.InlineKeyboardButton{{Text: file.Name, CallbackData: "file " + strconv.FormatInt(file.ID, 10)}})
		}

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      data.ChatID,
			Text:        "Files available for download",
			ReplyMarkup: kb,
		})
	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
			Text:   "unknown button",
		})
	}

}

func callbackFileHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	data := ctx.Value("data").(*pkg.TgData)
	user := ctx.Value("user").(*storage.User)
	l := ctx.Value("logger").(*zap.Logger)

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	id, err := strconv.Atoi(update.CallbackQuery.Data[5:])
	if err != nil {
		l.Error("failed to ReadAll", zap.Error(err))
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("Failed to Atoi")})
		return
	}

	file, err := storage.GetFile(db.DB, id)
	if err != nil {
		l.Error("failed to GetFile", zap.Error(err))
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("failed to GetFile")})
		return
	}
	raw, err := ipfs.Download(file.IpfsKey, "")
	if err != nil {
		l.Error("failed to Download", zap.Error(err))
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("failed to Download")})
		return
	}
	decrypted, err := dec(storage.User{
		ID:         user.ID,
		TgName:     "",
		PK:         user.PK,
		Department: user.Department,
		Level:      user.Level,
	}, string(raw))

	if err != nil {
		l.Error("failed to dec", zap.Error(err))
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("failed to dec")})
		return
	}

	m, err := b.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:   data.ChatID,
		Document: &models.InputFileUpload{Filename: file.Name, Data: bytes.NewReader([]byte(decrypted))},
		Caption:  "Document",
	})
	if err != nil {
		fmt.Println(m)
		l.Error("failed to send file", zap.Error(err))
		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: data.ChatID, Text: fmt.Sprintf("failed to send file")})
		return
	}
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
