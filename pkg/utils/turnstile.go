package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileError struct {
	Code string
}

func (e *TurnstileError) Error() string {
	return e.Code
}

type TurnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

type TurnstileRequest struct {
	Token  string `json:"response"`
	Secret string `json:"secret"`
}

func ValidateTurnstile(token, secret string) error {
	req := TurnstileRequest{Token: token, Secret: secret}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := http.Post(turnstileVerifyURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	var turnstileResp TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&turnstileResp); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK || !turnstileResp.Success {
		if len(turnstileResp.ErrorCodes) > 0 {
			return &TurnstileError{Code: turnstileResp.ErrorCodes[0]}
		}

		return &TurnstileError{Code: "unknown_error"}
	}

	return nil
}
