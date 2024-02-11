package main

import "github.com/go-playground/validator"

type RequestAdd struct {
	Level      int    `json:"level" validate:"required"`
	Department int    `json:"department" validate:"required"`
	Data       string `json:"data" validate:"required"`
}

type RequestCheck struct {
	Level         int    `json:"level" validate:"required"`
	Department    int    `json:"department" validate:"required"`
	WitLevel      string `json:"wit_level" validate:"required"`
	WitDepartment string `json:"wit_department" validate:"required"`
}

type CustomValidator struct {
	validator *validator.Validate
}

type Response struct {
	WitLevel string `json:"wit_level"`
	WitDep   string `json:"wit_dep"`
}
