package httpx

import (
	"encoding/json"
	"net/http"
	"strings"

	"back/internal/platform/errcode"
)

// HTTPError is a thin error wrapper intended for stable, user-facing messages.
type HTTPError string

func (e HTTPError) Error() string { return string(e) }

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteAPIError(w http.ResponseWriter, status int, code string, message string) {
	if code == "" {
		code = codeFromStatus(status)
	}
	if message == "" {
		message = http.StatusText(status)
	}
	WriteJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func WriteError(w http.ResponseWriter, status int, err error) {
	if code := errcode.FromError(err); code != "" {
		WriteAPIError(w, status, code, errcode.Message(code))
		return
	}
	WriteAPIError(w, status, codeFromStatus(status), err.Error())
}

func codeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusGatewayTimeout:
		return "UPSTREAM_TIMEOUT"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		if status >= 500 {
			return "INTERNAL_ERROR"
		}
		return "ERROR"
	}
}

func BearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
