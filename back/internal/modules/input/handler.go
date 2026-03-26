package input

import (
	"encoding/json"
	"net/http"

	"back/internal/platform/errcode"
	"back/internal/platform/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PostAnalyze(w http.ResponseWriter, r *http.Request) {
	var req ManualAnalyzeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	res, err := h.service.Analyze(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, errcode.HTTPStatus(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) PostCustom(w http.ResponseWriter, r *http.Request) {
	var req CustomProductRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	if err := h.service.CreateCustomProduct(r.Context(), req); err != nil {
		httpx.WriteError(w, errcode.HTTPStatus(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, CustomProductResponse{Status: "created"})
}
