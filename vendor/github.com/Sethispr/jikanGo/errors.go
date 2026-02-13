package jikan

import (
	"encoding/json"
	"net/http"
)

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Err     string `json:"error"`
}

func (e *Error) Error() string {
	if e.Err != "" {
		return e.Message + ": " + e.Err
	}
	return e.Message
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Status == t.Status
}

func (e *Error) IsNotFound() bool    { return e.Status == http.StatusNotFound }
func (e *Error) IsRateLimit() bool   { return e.Status == http.StatusTooManyRequests }
func (e *Error) IsServerError() bool { return e.Status >= 500 && e.Status < 600 }

func parseError(resp *http.Response) error {
	var e Error
	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		e.Message = http.StatusText(resp.StatusCode)
	}
	e.Status = resp.StatusCode
	return &e
}
