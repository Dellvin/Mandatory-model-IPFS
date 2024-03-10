package main

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"server/ipfs"

	"server/crypto"
	"server/security"
	"server/storage"
	"server/stribog"
)

func Encrypt(c echo.Context) error {
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

	base64Cipher, err := enc(storage.User{
		ID:         req.ID,
		TgName:     "",
		PK:         req.PK,
		Department: req.Department,
		Level:      req.Level,
	}, req.File)
	if err != nil {
		c.Logger().Errorf("failed to enc: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	link, err := ipfs.Upload("", []byte(base64Cipher))
	if err != nil {
		c.Logger().Errorf("failed to Upload: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if err = storage.AddFile(db.DB, storage.File{
		Name:    req.File,
		IpfsKey: link,
		UserID:  req.ID,
	}); err != nil {
		c.Logger().Errorf("failed to AddFile: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ResponseFile{File: base64Cipher})
}

func enc(user storage.User, file string) (string, error) {
	accum, err := storage.GetWitness(db.DB, user.PK)
	if err != nil {
		return "", fmt.Errorf("failed to GetWitness: %s", err.Error())
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString wit level: %s", err.Error())
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString wit dep: %s", err.Error())
	}

	if err = security.Check(user.Level, user.Department, witLevel, witDep); err != nil {
		return "", fmt.Errorf("failed to Check: %s", err.Error())
	}

	cipherRaw, depAuthRaw, levelAuthRaw, err := crypto.Encrypt(user.Department, user.Level, []byte(file))
	if err != nil {
		return "", fmt.Errorf("failed to Encrypt: %s", err.Error())
	}

	base64Cipher := base64.StdEncoding.EncodeToString(cipherRaw)
	if err = storage.SetAbe(db.DB, storage.AbeAuth{
		ID:        base64.StdEncoding.EncodeToString(stribog.New512().Sum([]byte(base64Cipher))),
		LevelAuth: base64.StdEncoding.EncodeToString(levelAuthRaw),
		DepAuth:   base64.StdEncoding.EncodeToString(depAuthRaw)}); err != nil {
		return "", fmt.Errorf("failed to SetAbe: %s", err.Error())
	}

	return base64Cipher, nil
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

	file, err := storage.GetFile(db.DB, req.ID)
	if err != nil {
		c.Logger().Errorf("failed to GetFile: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	raw, err := ipfs.Download(file.IpfsKey, "")
	if err != nil {
		c.Logger().Errorf("failed to Download: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	decrypted, err := dec(storage.User{
		ID:         req.ID,
		TgName:     "",
		PK:         req.PK,
		Department: req.Department,
		Level:      req.Level,
	}, string(raw))

	if err != nil {
		c.Logger().Errorf("failed to dec: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ResponseFile{File: string(decrypted)}) // TODO add base64
}

func dec(user storage.User, file string) (string, error) {
	if err := storage.CheckUserPK(db.DB, user.ID, user.PK); err != nil {
		return "", fmt.Errorf("failed to CheckUserPK: %s", err.Error())
	}

	accum, err := storage.GetWitness(db.DB, user.PK)
	if err != nil {
		return "", fmt.Errorf("failed to GetWitness: %s", err.Error())
	}

	witLevel, err := base64.StdEncoding.DecodeString(accum.WitnessLevel)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString wit level: %s", err.Error())
	}

	witDep, err := base64.StdEncoding.DecodeString(accum.WitnessDep)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString wit dep: %s", err.Error())
	}

	if err = security.Check(user.Level, user.Department, witLevel, witDep); err != nil {
		return "", fmt.Errorf("failed to Check: %s", err.Error())
	}

	auth, err := storage.GetAbe(db.DB, base64.StdEncoding.EncodeToString(stribog.New512().Sum([]byte(file))))
	if err != nil {
		return "", fmt.Errorf("failed to GetAbe: %s", err.Error())
	}

	authLevel, err := base64.StdEncoding.DecodeString(auth.LevelAuth)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString wit auth: %s", err.Error())
	}

	authDep, err := base64.StdEncoding.DecodeString(auth.DepAuth)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString dep auth: %s", err.Error())
	}

	cipher, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return "", fmt.Errorf("failed to DecodeString file: %s", err.Error())
	}

	decrypted, err := crypto.Decrypt(user.Department, user.Level, cipher, authDep, authLevel)
	if err != nil {
		return "", fmt.Errorf("failed to Encrypt: %s", err.Error())
	}

	return string(decrypted), nil
}
