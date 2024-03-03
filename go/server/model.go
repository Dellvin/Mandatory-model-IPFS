package main

import "github.com/go-playground/validator"

type RequestAdd struct {
	Level      int    `json:"Level"`
	Department int    `json:"Department"`
	ID         int    `json:"ID"`
	PK         string `json:"PK"`
}

type RequestFile struct {
	Level      int    `json:"Level" validate:"required"`
	Department int    `json:"Department" validate:"required"`
	ID         int    `json:"ID" validate:"required"`
	PK         string `json:"PK"`
	File       string `json:"File" validate:"required"`
}

type ResponseFile struct {
	File string `json:"File" validate:"required"`
}

type CustomValidator struct {
	validator *validator.Validate
}
