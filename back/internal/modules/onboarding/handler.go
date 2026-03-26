package onboarding

import (
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

func (h *Handler) GetV1Options(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, h.service.Options())
}

func (h *Handler) GetV1Status(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	status, err := h.service.Status(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, StatusResponse{ProfileCompleted: status.ProfileCompleted})
}
