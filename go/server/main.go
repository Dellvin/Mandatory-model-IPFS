package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/jackc/pgx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"server/config"
	"server/storage"
)

var db storage.Database

const adminHeader = "admin"

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

	// Routes
	e.POST("/admin/add", add)
	e.PUT("/admin/check", check)
	e.DELETE("/admin/delete", delete)
	e.GET("/admin/all", getAll)
	e.POST("/file/encrypt", encrypt)
	e.POST("/file/decrypt", decrypt)

	// Start server
	e.Logger.Fatal(e.Start(cfg.Server.Port))
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
