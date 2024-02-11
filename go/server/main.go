package main

import (
	"encoding/base64"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"server/security"
)

func main() {
	// Echo instance
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/accumulator", add)
	e.PUT("/accumulator", check)
	e.DELETE("/accumulator", delete)

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
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString data: %w", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	witLevel, witDep, err := security.Add(req.Level, req.Department, data)
	if err != nil {
		c.Logger().Errorf("failed to Add: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, Response{WitLevel: base64.StdEncoding.EncodeToString(witLevel), WitDep: base64.StdEncoding.EncodeToString(witDep)})
}

// Handler
func check(c echo.Context) error {
	var req RequestCheck
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
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	err := security.Delete(req.Level, req.Department, []byte(req.Data))
	if err != nil {
		c.Logger().Errorf("failed to Delete: %w", err)
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
