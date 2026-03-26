package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"back/internal/platform/authctx"
	"back/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.Register(r.Context(), input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"status": "verification_sent"})
}

func (h *Handler) PostSendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.SendVerificationCode(r.Context(), body.Email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "verification_sent"})
}

func (h *Handler) PostVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input VerifyEmailInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.VerifyEmail(r.Context(), input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "email_verified"})
}

func (h *Handler) PostLogin(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.Login(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostSendLoginCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.SendLoginCode(r.Context(), body.Email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "login_code_sent"})
}

func (h *Handler) PostSendPasswordResetCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.SendPasswordResetCode(r.Context(), body.Email); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "password_reset_code_sent"})
}

func (h *Handler) PostLoginWithCode(w http.ResponseWriter, r *http.Request) {
	var input EmailCodeLoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.LoginWithCode(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostResetPassword(w http.ResponseWriter, r *http.Request) {
	var input ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.ResetPassword(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostGoogle(w http.ResponseWriter, r *http.Request) {
	var input GoogleLoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.GoogleLogin(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.service.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostSetPassword(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var input SetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.service.SetPassword(r.Context(), userID, input); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteAPIError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "password_set"})
}
