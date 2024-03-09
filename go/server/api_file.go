package main

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"net/http"
	"server/config"
	"server/pkg"
	"strings"

	"github.com/labstack/echo/v4"

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

func upload(key string, config config.Config, file []byte) (string, error) {
	//API = "ipfs.io"
	fmt.Println("Initializing ipfs node...")
	n, err := ipfssenc.GetRWIPFSNode(config.IPFS.API)
	if err != nil {
		return "", err
	}
	if !n.IsUp() {
		return "", pkg.ErrNoIPFS
	}

	link, err := ipfssenc.Put(n, bytes.NewReader(file))
	if err != nil {
		return "", fmt.Errorf("failed to Put: %w", err)
	}

	l := string(link)
	if !strings.HasPrefix(l, "/ipfs/") {
		l = "/ipfs/" + l
	}
	return base32.StdEncoding.EncodeToString([]byte(key)), nil
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
