package recipe

import (
	"context"
	"errors"
	"strings"

	"back/internal/modules/input"
	"back/internal/modules/nutrition"
	"back/internal/platform/errcode"
)

type ManualAnalyzer interface {
	Analyze(ctx context.Context, req input.ManualAnalyzeRequest) (input.ManualAnalyzeResponse, error)
}

type Service struct {
	manual ManualAnalyzer
}

func NewService(manual ManualAnalyzer) *Service {
	return &Service{manual: manual}
}

func (s *Service) Analyze(ctx context.Context, req AnalyzeRequest) (AnalyzeResponse, error) {
	if len(req.Ingredients) == 0 {
		return AnalyzeResponse{}, errcode.EmptyRecipe
	}

	merged := make(map[string]float64, len(req.Ingredients))
	order := make([]string, 0, len(req.Ingredients))
	for _, ingredient := range req.Ingredients {
		if input.NormalizeName(ingredient.Name) == "" {
			return AnalyzeResponse{}, errcode.InvalidName
		}
		if !input.IsValidAmount(ingredient.Amount) {
			return AnalyzeResponse{}, errcode.InvalidAmount
		}

		name := input.NormalizeName(ingredient.Name)
		if _, ok := merged[name]; !ok {
			order = append(order, name)
		}
		merged[name] += ingredient.Amount
	}

	resolved := make([]ResolvedIngredient, 0, len(order))
	total := nutrition.Nutrients{}
	totalAmount := 0.0

	for _, name := range order {
		amount := merged[name]
		result, err := s.manual.Analyze(ctx, input.ManualAnalyzeRequest{
			Name:   name,
			Amount: amount,
		})
		if err != nil {
			if errors.Is(err, errcode.NotFound) {
				return AnalyzeResponse{}, errcode.IngredientNotFound
			}
			return AnalyzeResponse{}, err
		}
		if result.Product == nil {
			return AnalyzeResponse{}, errcode.IngredientNotFound
		}

		resolved = append(resolved, ResolvedIngredient{
			Name:     result.Product.Name,
			Amount:   amount,
			Calories: result.Product.Calories,
			Protein:  result.Product.Protein,
			Fat:      result.Product.Fat,
			Carbs:    result.Product.Carbs,
		})

		total = nutrition.Add(total, nutrition.Nutrients{
			Calories: result.Product.Calories,
			Protein:  result.Product.Protein,
			Fat:      result.Product.Fat,
			Carbs:    result.Product.Carbs,
		})
		totalAmount += amount
	}

	return AnalyzeResponse{
		Product: input.ProductResult{
			Name:     strings.TrimSpace(req.Name),
			Calories: total.Calories,
			Protein:  total.Protein,
			Fat:      total.Fat,
			Carbs:    total.Carbs,
		},
		Per100G: func() input.ProductResult {
			per100g, err := nutrition.Per100gFromTotal(total, totalAmount)
			if err != nil {
				return input.ProductResult{}
			}
			return input.ProductResult{
				Name:     strings.TrimSpace(req.Name),
				Calories: per100g.Calories,
				Protein:  per100g.Protein,
				Fat:      per100g.Fat,
				Carbs:    per100g.Carbs,
			}
		}(),
		Ingredients: resolved,
		Confidence:  1.0,
		AmountG:     totalAmount,
	}, nil
}
