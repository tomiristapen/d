package user

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

func (h *Handler) DeleteV1User(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	if err := h.service.Delete(r.Context(), userID); err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.HTTPError("user not found"))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

