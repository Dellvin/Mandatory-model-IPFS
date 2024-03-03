package main

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"server/security"
	"server/storage"
)

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

	user, err := storage.AddUser(db.DB, req.Level, req.Department)
	if err != nil {
		return fmt.Errorf("failed to AddUser: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(user.PK)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString data: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	witLevel, witDep, err := security.Add(req.Level, req.Department, data)
	if err != nil {
		c.Logger().Errorf("failed to Add: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err := storage.SetWitness(db.DB, storage.Witness{ID: user.PK, WitnessLevel: base64.StdEncoding.EncodeToString(witLevel), WitnessDep: base64.StdEncoding.EncodeToString(witDep)}); err != nil {
		c.Logger().Errorf("failed to SetWitness level: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, user)
}

// Handler
func check(c echo.Context) error {
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

	if err := storage.CheckUser(db.DB, req.ID, req.PK); err != nil {
		return fmt.Errorf("failed to CheckUser: %w", err)
	}

	accum, err := storage.GetWitness(db.DB, req.PK)
	if err != nil {
		c.Logger().Errorf("failed to GetWitness: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit level: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit dep: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	if err = security.Check(req.Level, req.Department, witLevel, witDep); err != nil {
		c.Logger().Errorf("failed to Delete: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

// Handler
func delete(c echo.Context) error { // TODO add delete from DB
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

	if err := storage.DeleteUser(db.DB, req.ID, req.PK); err != nil {
		return fmt.Errorf("failed to DeleteUser: %w", err)
	}

	if err := storage.DeleteWitness(db.DB, req.PK); err != nil {
		return fmt.Errorf("failed to DeleteWitness: %w", err)
	}

	err := security.Delete(req.Level, req.Department, []byte(req.PK))
	if err != nil {
		c.Logger().Errorf("failed to Delete: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, http.StatusText(http.StatusOK))
}

// Handler
func getAll(c echo.Context) error {
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

	usrs, err := storage.GetAll(db.DB)
	if err != nil {
		return fmt.Errorf("failed to GetAll: %w", err)
	}

	return c.JSON(http.StatusOK, usrs)
}
