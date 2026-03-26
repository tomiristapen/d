package ingredient

import (
	"net/http"

	"back/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAutocomplete(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Autocomplete(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, AutocompleteResponse{Items: items})
}
