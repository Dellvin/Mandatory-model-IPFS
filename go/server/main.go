package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-playground/validator"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"server/config"
	"server/storage"
)

var db storage.Database

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}

func main() {
	cfgPath, err := config.ParseFlags()
	if err != nil {
		panic(err)
	}

	fmt.Println("-------------:", cfgPath)
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		panic(err)
	}

	if err = db.Init(*cfg); err != nil {
		panic(err)
	}

	if err = storage.CreateTableAbe(db.DB); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(err)
	}

	if err = storage.CreateTableWitness(db.DB); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(err)
	}

	if err = storage.CreateTableAccumulator(db.DB); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(err)
	}

	if err = storage.CreateTableUser(db.DB); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		panic(err)
	}

	// Echo instance
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(cfg.Process)

	// Routes
	e.POST("/admin/add", add)
	e.PUT("/admin/check", check)
	e.DELETE("/admin/delete", delete)
	e.GET("/admin/all", getAll)
	e.POST("/file/encrypt", encrypt)
	e.POST("/file/decrypt", decrypt)

	// set up tg bot
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithMiddlewares(showMessageWithUserName),
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New("6893355444:AAG0A2AJ3GjcJ6eyf9u456YyZSFJFZ_ADEk", opts...)
	if nil != err {
		// panics for the sake of simplicity.
		// you should handle this error properly in your code.
		panic(err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// Start server
		return fmt.Errorf("failed to Start server: %w", e.Start(cfg.Server.Port))
	})
	g.Go(func() error {
		// Start tg bot listener
		b.Start(ctx)
		return fmt.Errorf("failed to Start tg bot")
	})

	if err = g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func showMessageWithUserName(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		l := zap.L().With(zap.String("username", update.Message.From.Username), zap.String("message", update.Message.Text))
		if user, err := storage.GetUserByTgName(db.DB, update.Message.From.Username); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				l.Info("unknown user")
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Your account not found in database! Please contact your system administrator.",
				})
				if err != nil {
					l.Error("failed to send message", zap.Error(err))
				}
				return
			} else {
				l.Error("failed to GetUserByTgName", zap.Error(err))
				_, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Internal server error. Sorry",
				})
				if err != nil {
					l.Error("failed to send message", zap.Error(err))
				}
				return
			}
		} else {
			l.Info("found user")
			ctx = context.WithValue(ctx, "user", &user)
		}

		ctx = context.WithValue(ctx, "logger", l)

		next(ctx, b, update)
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
