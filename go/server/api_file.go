package main

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"

	"server/crypto"
	"server/security"
	"server/storage"
	"server/stribog"
)

func encrypt(c echo.Context) error {
	var req RequestFile

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := storage.CheckUserPK(db.DB, req.ID, req.PK); err != nil {
		c.Logger().Errorf("failed to CheckUserPK: %s", err.Error())
		return c.JSON(http.StatusForbidden, err.Error())
	}

	accum, err := storage.GetWitness(db.DB, req.PK)
	if err != nil {
		c.Logger().Errorf("failed to GetWitness: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit level: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit dep: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if err = security.Check(req.Level, req.Department, witLevel, witDep); err != nil {
		c.Logger().Errorf("failed to Check: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	cipherRaw, depAuthRaw, levelAuthRaw, err := crypto.Encrypt(req.Department, req.Level, []byte(req.File))
	if err != nil {
		c.Logger().Errorf("failed to Encrypt: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	base64Cipher := base64.StdEncoding.EncodeToString(cipherRaw)
	if err = storage.SetAbe(db.DB, storage.AbeAuth{
		ID:        base64.StdEncoding.EncodeToString(stribog.New512().Sum([]byte(base64Cipher))),
		LevelAuth: base64.StdEncoding.EncodeToString(levelAuthRaw),
		DepAuth:   base64.StdEncoding.EncodeToString(depAuthRaw)}); err != nil {
		c.Logger().Errorf("failed to SetAbe: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ResponseFile{File: base64Cipher})
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

	if err := storage.CheckUserPK(db.DB, req.ID, req.PK); err != nil {
		c.Logger().Errorf("failed to CheckUserPK: %s", err.Error())
		return c.JSON(http.StatusForbidden, err.Error())
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
		c.Logger().Errorf("failed to Check: %s", err.Error())
		return c.JSON(http.StatusForbidden, err.Error())
	}

	auth, err := storage.GetAbe(db.DB, base64.StdEncoding.EncodeToString(stribog.New512().Sum([]byte(req.File))))
	if err != nil {
		c.Logger().Errorf("failed to GetAbe: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	authLevel, err := base64.StdEncoding.DecodeString(auth.LevelAuth)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString wit auth: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	authDep, err := base64.StdEncoding.DecodeString(auth.DepAuth)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString dep auth: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	cipher, err := base64.StdEncoding.DecodeString(req.File)
	if err != nil {
		c.Logger().Errorf("failed to DecodeString file: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	file, err := crypto.Decrypt(req.Department, req.Level, cipher, authDep, authLevel)
	if err != nil {
		c.Logger().Errorf("failed to Encrypt: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ResponseFile{File: string(file)}) // TODO add base64
}
