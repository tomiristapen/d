package input

import (
	"context"
	"errors"

	"back/internal/modules/nutrition"
	"back/internal/modules/product"
	"back/internal/platform/errcode"
	"back/internal/platform/textnorm"
)

const singleMatchConfidence = 0.85

type ProductService interface {
	FindExactBaseProduct(ctx context.Context, name string) (product.BaseProduct, error)
	SuggestBaseProducts(ctx context.Context, name string) ([]product.BaseProduct, error)
	CreateCustomProduct(ctx context.Context, req product.CustomProductRequest) (product.BaseProduct, error)
}

type Service struct {
	products ProductService
}

const MaxAmountG = 2000.0

func NewService(products ProductService) *Service {
	return &Service{products: products}
}

func NormalizeName(name string) string {
	return textnorm.LowerTrim(name)
}

func IsValidAmount(amount float64) bool {
	return amount > 0 && amount <= MaxAmountG
}

func (s *Service) Analyze(ctx context.Context, req ManualAnalyzeRequest) (ManualAnalyzeResponse, error) {
	if NormalizeName(req.Name) == "" {
		return ManualAnalyzeResponse{}, errcode.InvalidName
	}
	if !IsValidAmount(req.Amount) {
		return ManualAnalyzeResponse{}, errcode.InvalidAmount
	}

	name := NormalizeName(req.Name)
	item, err := s.products.FindExactBaseProduct(ctx, name)
	if err == nil {
		return buildResolvedResponse(item, req.Amount, 1.0)
	}
	if !errors.Is(err, errcode.NotFound) {
		return ManualAnalyzeResponse{}, err
	}

	suggestions, err := s.products.SuggestBaseProducts(ctx, name)
	if err != nil {
		return ManualAnalyzeResponse{}, err
	}
	suggestions = dedupeSuggestions(suggestions)
	if len(suggestions) == 1 {
		return buildResolvedResponse(suggestions[0], req.Amount, singleMatchConfidence)
	}

	out := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		out = append(out, suggestion.Name)
	}
	if len(out) == 0 {
		return ManualAnalyzeResponse{}, errcode.NotFound
	}

	return ManualAnalyzeResponse{
		Product:     nil,
		Suggestions: out,
		Confidence:  nil,
		AmountG:     req.Amount,
	}, nil
}

func (s *Service) CreateCustomProduct(ctx context.Context, req CustomProductRequest) error {
	_, err := s.products.CreateCustomProduct(ctx, product.CustomProductRequest{
		Name:     req.Name,
		Calories: req.Calories,
		Protein:  req.Protein,
		Fat:      req.Fat,
		Carbs:    req.Carbs,
	})
	return err
}

func buildResolvedResponse(item product.BaseProduct, amount float64, confidence float64) (ManualAnalyzeResponse, error) {
	scaled, err := nutrition.ScalePer100g(nutrition.Nutrients{
		Calories: item.Calories,
		Protein:  item.Protein,
		Fat:      item.Fat,
		Carbs:    item.Carbs,
	}, amount)
	if err != nil {
		return ManualAnalyzeResponse{}, err
	}

	return ManualAnalyzeResponse{
		Product: &ProductResult{
			Name:     item.Name,
			Calories: scaled.Calories,
			Protein:  scaled.Protein,
			Fat:      scaled.Fat,
			Carbs:    scaled.Carbs,
		},
		Per100G: &ProductResult{
			Name:     item.Name,
			Calories: item.Calories,
			Protein:  item.Protein,
			Fat:      item.Fat,
			Carbs:    item.Carbs,
		},
		Suggestions: []string{},
		Confidence:  &confidence,
		AmountG:     amount,
	}, nil
}

func dedupeSuggestions(items []product.BaseProduct) []product.BaseProduct {
	if len(items) < 2 {
		return items
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]product.BaseProduct, 0, len(items))
	for _, item := range items {
		key := NormalizeName(item.Name)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
