package diary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"back/internal/modules/input"
	"back/internal/modules/nutrition"
	"back/internal/modules/recipe"
	"back/internal/platform/authctx"
	"back/internal/platform/errcode"
	"back/internal/platform/httpx"
)

const timezoneOffsetHeader = "X-Timezone-Offset-Minutes"

type ManualAnalyzer interface {
	Analyze(ctx context.Context, req input.ManualAnalyzeRequest) (input.ManualAnalyzeResponse, error)
}

type RecipeAnalyzer interface {
	Analyze(ctx context.Context, req recipe.AnalyzeRequest) (recipe.AnalyzeResponse, error)
}

type Handler struct {
	service *Service
	manual  ManualAnalyzer
	recipe  RecipeAnalyzer
}

func NewHandler(service *Service, manual ManualAnalyzer, recipe RecipeAnalyzer) *Handler {
	return &Handler{service: service, manual: manual, recipe: recipe}
}

func (h *Handler) PostEntry(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req AddToDiaryInput
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	offsetMinutes, err := timezoneOffsetFromRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	saved, err := h.service.AddToDiary(r.Context(), AddToDiaryInput{
		UserID:                userID,
		Source:                req.Source,
		Name:                  req.Name,
		AmountG:               req.AmountG,
		Per100G:               req.Per100G,
		Ingredients:           req.Ingredients,
		TimezoneOffsetMinutes: offsetMinutes,
		IdempotencyKey:        strings.TrimSpace(r.Header.Get("Idempotency-Key")),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, saved)
}

func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	offsetMinutes, err := timezoneOffsetFromRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := h.service.GetToday(r.Context(), userID, offsetMinutes)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrTargetsNotFound) {
			status = http.StatusNotFound
		}
		httpx.WriteError(w, status, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) PostManualAddToDiary(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req input.ManualAnalyzeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	res, err := h.manual.Analyze(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if res.Product == nil {
		httpx.WriteError(w, http.StatusBadRequest, errcode.NotFound)
		return
	}

	offsetMinutes, err := timezoneOffsetFromRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	per100g, err := nutrition.Per100gFromTotal(nutrition.Nutrients{
		Calories: res.Product.Calories,
		Protein:  res.Product.Protein,
		Fat:      res.Product.Fat,
		Carbs:    res.Product.Carbs,
	}, res.AmountG)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	saved, err := h.service.AddToDiary(r.Context(), AddToDiaryInput{
		UserID:                userID,
		Source:                "manual",
		Name:                  res.Product.Name,
		AmountG:               res.AmountG,
		Per100G:               per100g,
		Ingredients:           []string{res.Product.Name},
		TimezoneOffsetMinutes: offsetMinutes,
		IdempotencyKey:        strings.TrimSpace(r.Header.Get("Idempotency-Key")),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, saved)
}

func (h *Handler) PostRecipeAddToDiary(w http.ResponseWriter, r *http.Request) {
	userID := authctx.UserID(r.Context())
	if userID == "" {
		httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req recipe.AnalyzeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.HTTPError("invalid json body"))
		return
	}

	res, err := h.recipe.Analyze(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	offsetMinutes, err := timezoneOffsetFromRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	per100g, err := nutrition.Per100gFromTotal(nutrition.Nutrients{
		Calories: res.Product.Calories,
		Protein:  res.Product.Protein,
		Fat:      res.Product.Fat,
		Carbs:    res.Product.Carbs,
	}, res.AmountG)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	saved, err := h.service.AddToDiary(r.Context(), AddToDiaryInput{
		UserID:                userID,
		Source:                "recipe",
		Name:                  res.Product.Name,
		AmountG:               res.AmountG,
		Per100G:               per100g,
		Ingredients:           recipeIngredientNames(res.Ingredients),
		TimezoneOffsetMinutes: offsetMinutes,
		IdempotencyKey:        strings.TrimSpace(r.Header.Get("Idempotency-Key")),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, saved)
}

func recipeIngredientNames(items []recipe.ResolvedIngredient) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		names = append(names, item.Name)
	}
	return names
}

func timezoneOffsetFromRequest(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.Header.Get(timezoneOffsetHeader))
	if raw == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(raw)
	if err != nil {
		return 0, httpx.HTTPError("invalid timezone offset")
	}
	return offset, nil
}
