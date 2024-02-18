package main

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"server/abe"
	"server/config"
	"server/security"
	"server/storage"
	"server/stribog"
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

	if err = storage.CreateTableAbe(db.DB); err != nil {
		panic(err)
	}

	if err = storage.CreateTableAccum(db.DB); err != nil {
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
	e.POST("/file/encrypt", encrypt)
	e.POST("/file/decrypt", decrypt)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// Handler
func add(c echo.Context) error {
	var req RequestAdd
	if c.Request().Header.Get("X-Admin-Key") != adminHeader {
		c.Logger().Errorf("incorrect admin header")
		return echo.NewHTTPError(http.StatusForbidden, "incorrect admin header")
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(req.ID)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString data: %w", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	witLevel, witDep, err := security.Add(req.Level, req.Department, data)
	if err != nil {
		c.Logger().Errorf("failed to Add: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err := storage.SetAccum(db.DB, storage.Accum{ID: req.ID, WitnessLevel: base64.StdEncoding.EncodeToString(witLevel), WitnessDep: base64.StdEncoding.EncodeToString(witDep)}); err != nil {
		c.Logger().Errorf("failed to SetAccum level: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

// Handler
func check(c echo.Context) error {
	var req RequestCheck
	if c.Request().Header.Get("X-Admin-Key") != adminHeader {
		c.Logger().Errorf("incorrect admin header")
		return echo.NewHTTPError(http.StatusForbidden, "incorrect admin header")
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	witLevel, err := base64.StdEncoding.DecodeString(req.WitLevel)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit_level: %w", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	witDep, err := base64.StdEncoding.DecodeString(req.WitDepartment)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit_department: %w", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err = security.Check(req.Level, req.Department, witLevel, witDep); err != nil {
		c.Logger().Errorf("failed to Check: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

// Handler
func delete(c echo.Context) error {
	var req RequestAdd
	if c.Request().Header.Get("X-Admin-Key") != adminHeader {
		c.Logger().Errorf("incorrect admin header")
		return echo.NewHTTPError(http.StatusForbidden, "incorrect admin header")
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	err := security.Delete(req.Level, req.Department, []byte(req.ID))
	if err != nil {
		c.Logger().Errorf("failed to Delete: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

// Handler
func encrypt(c echo.Context) error {
	var req RequestFile

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	accum, err := storage.GetWitness(db.DB, req.ID)
	if err != nil {
		c.Logger().Errorf("failed to GetWitness: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit level: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit dep: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err = security.Check(req.Level, req.Department, witLevel, witDep); err != nil {
		c.Logger().Errorf("failed to Delete: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	cipherRaw, depAuthRaw, levelAuthRaw, err := abe.Encrypt(req.Department, req.Level, []byte(req.File))
	if err != nil {
		c.Logger().Errorf("failed to Encrypt: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err = storage.SetAbe(db.DB, storage.AbeAuth{ID: string(stribog.New512().Sum(cipherRaw)), LevelAuth: base64.StdEncoding.EncodeToString(levelAuthRaw), DepAuth: base64.StdEncoding.EncodeToString(depAuthRaw)}); err != nil {
		c.Logger().Errorf("failed to SetAbe: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ResponseFile{File: base64.StdEncoding.EncodeToString(cipherRaw)})
}

// Handler
func decrypt(c echo.Context) error {
	var req RequestFile

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	accum, err := storage.GetWitness(db.DB, req.ID)
	if err != nil {
		c.Logger().Errorf("failed to GetWitness: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit level: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit dep: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err = security.Check(req.Level, req.Department, witLevel, witDep); err != nil {
		c.Logger().Errorf("failed to Delete: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	auth, err := storage.GetAbe(db.DB, string(stribog.New512().Sum([]byte(req.File))))
	if err != nil {
		c.Logger().Errorf("failed to GetAbe: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	authLevel, err := base64.StdEncoding.DecodeString(auth.LevelAuth)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit auth: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	authDep, err := base64.StdEncoding.DecodeString(auth.DepAuth)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString dep auth: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	file, err := abe.Decrypt(req.Department, req.Level, []byte(req.File), authDep, authLevel)
	if err != nil {
		c.Logger().Errorf("failed to Encrypt: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ResponseFile{File: base64.StdEncoding.EncodeToString(file)})
}
