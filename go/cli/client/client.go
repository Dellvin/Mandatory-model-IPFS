package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type requestAdd struct {
	Level      int    `json:"level" validate:"required"`
	Department int    `json:"department" validate:"required"`
	Data       string `json:"data" validate:"required"`
}

type requestCheck struct {
	Level         int    `json:"level" validate:"required"`
	Department    int    `json:"department" validate:"required"`
	WitLevel      string `json:"wit_level" validate:"required"`
	WitDepartment string `json:"wit_department" validate:"required"`
}

type Response struct {
	WitLevel string `json:"wit_level"`
	WitDep   string `json:"wit_dep"`
}

func Add(level, dep int, data string) (string, string, error) {
	body, err := json.Marshal(requestAdd{Level: level, Department: dep, Data: data})
	if err != nil {
		return "", "", fmt.Errorf("failed to Marshal: %w", err)
	}
	r, err := http.NewRequest("POST", "127.0.0.1:1323/accumulator", bytes.NewBuffer(body))
	if err != nil {
		return "", "", fmt.Errorf("failed to NewRequest: %w", err)
	}

	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		return "", "", fmt.Errorf("failed to Do: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bad status: %d, %s", res.StatusCode, res.Status)
	}

	var resp Response
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", "", fmt.Errorf("failed to Unmarshal: %w", err)
	}

	return resp.WitLevel, resp.WitDep, nil
}

func Delete(level, dep int, data string) error {
	body, err := json.Marshal(requestAdd{Level: level, Department: dep, Data: data})
	if err != nil {
		return fmt.Errorf("failed to Marshal: %w", err)
	}
	r, err := http.NewRequest("DELETE", "127.0.0.1:1323/accumulator", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to NewRequest: %w", err)
	}

	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to Do: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d, %s", res.StatusCode, res.Status)
	}

	return nil
}

func Check(level, dep int, witLevel, witDep string) error {
	body, err := json.Marshal(requestCheck{Level: level, Department: dep, WitLevel: witLevel, WitDepartment: witDep})
	if err != nil {
		return fmt.Errorf("failed to Marshal: %w", err)
	}
	r, err := http.NewRequest("PUT", "127.0.0.1:1323/accumulator", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to NewRequest: %w", err)
	}

	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to Do: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d, %s", res.StatusCode, res.Status)
	}

	return nil
}
