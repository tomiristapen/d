package profile

import (
	"encoding/json"
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

func (h *Handler) GetV1Profile(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	p, err := h.service.Get(r.Context(), userID)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.HTTPError("profile not found"))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) PutV1Profile(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	userID := authctx.UserID(r.Context())
	p, err := h.service.Update(r.Context(), userID, req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toProfileResponse(p))
}

func (h *Handler) PostV1Reset(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if err := h.service.Reset(r.Context(), userID); err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.HTTPError("profile not found"))
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ResetResponse{Status: "profile_reset"})
}

func toProfileResponse(p Profile) ProfileResponse {
	return ProfileResponse{
		UserID:               p.UserID,
		Age:                  p.Age,
		Gender:               p.Gender,
		HeightCM:             p.HeightCM,
		WeightKG:             p.WeightKG,
		ActivityLevel:        p.ActivityLevel,
		Goal:                 p.Goal,
		Allergies:            p.Allergies,
		CustomAllergies:      p.CustomAllergies,
		Intolerances:         p.Intolerances,
		DietType:             p.DietType,
		ReligiousRestriction: p.ReligiousRestriction,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}
