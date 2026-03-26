package product

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"back/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := normalizeBarcode(chi.URLParam(r, "barcode"))
	if !isValidBarcode(barcode) {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid barcode"))
		return
	}

	p, err := h.service.LookupByBarcode(r.Context(), barcode)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.HTTPError("product not found"))
			return
		}
		if isTimeoutError(err) {
			httpx.WriteError(w, http.StatusGatewayTimeout, httpx.HTTPError("upstream timeout"))
			return
		}
		httpx.WriteError(w, http.StatusBadGateway, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) PostOCRDraft(w http.ResponseWriter, r *http.Request) {
	var req OCRDraftRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 25<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	draft, err := h.service.BuildOCRDraft(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidOCRRequest) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError(err.Error()))
			return
		}
		httpx.WriteError(w, http.StatusBadGateway, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, draft)
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}
