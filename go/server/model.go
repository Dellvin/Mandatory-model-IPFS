package main

import "github.com/go-playground/validator"

type RequestAdd struct {
	Level      int    `json:"level" validate:"required"`
	Department int    `json:"department" validate:"required"`
	ID         string `json:"ID" validate:"required"`
}

type RequestCheck struct {
	Level         int    `json:"level" validate:"required"`
	Department    int    `json:"department" validate:"required"`
	WitLevel      string `json:"wit_level" validate:"required"`
	WitDepartment string `json:"wit_department" validate:"required"`
}

type RequestFile struct {
	Level      int    `json:"level" validate:"required"`
	Department int    `json:"department" validate:"required"`
	ID         string `json:"ID" validate:"required"`
	File       string `json:"file" validate:"required"`
}

type ResponseFile struct {
	File string `json:"file" validate:"required"`
}

type CustomValidator struct {
	validator *validator.Validate
}
