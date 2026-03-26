package product

import (
	"context"
	"fmt"
	"log"
	"strings"

	"back/internal/platform/errcode"
	"back/internal/platform/textnorm"
)

type CustomProductRequest struct {
	Name     string  `json:"name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type Service struct {
	repo            Repository
	provider        ExternalProvider
	baseProducts    BaseProductRepository
	ocrMode         string
	ocrDraftBuilder *OCRDraftBuilder
}

func NewService(repo Repository, provider ExternalProvider, baseProducts BaseProductRepository) *Service {
	if baseProducts == nil {
		baseProducts = NewNoopBaseProductRepository()
	}

	s := &Service{
		repo:         repo,
		provider:     provider,
		baseProducts: baseProducts,
	}

	ocrClient := OCRClient(NewStubOCRClient())
	s.ocrMode = "stub"
	if tc, err := NewTesseractOCRClient(); err == nil {
		ocrClient = tc
		s.ocrMode = "tesseract"
	} else {
		log.Printf("product: OCR fallback (stub): %v", err)
	}
	s.ocrDraftBuilder = NewOCRDraftBuilder(ocrClient, baseProducts)
	return s
}

func (s *Service) OCRMode() string {
	if s == nil {
		return ""
	}
	return s.ocrMode
}

func (s *Service) BuildOCRDraft(ctx context.Context, req OCRDraftRequest) (OCRDraftDTO, error) {
	if s == nil || s.ocrDraftBuilder == nil {
		return OCRDraftDTO{}, fmt.Errorf("%w: ocr draft not configured", ErrInvalidOCRRequest)
	}
	return s.ocrDraftBuilder.Build(ctx, req, s.ocrMode)
}

func (s *Service) LookupByBarcode(ctx context.Context, barcode string) (Product, error) {
	candidates := barcodeLookupCandidates(barcode)
	if len(candidates) == 0 {
		return Product{}, ErrNotFound
	}

	for _, candidate := range candidates {
		p, err := s.repo.GetByBarcode(ctx, candidate)
		if err == nil {
			p.Source = "cache"
			return p, nil
		}
		if err != ErrNotFound {
			return Product{}, err
		}
	}

	for _, candidate := range candidates {
		fresh, err := s.provider.FetchByBarcode(ctx, candidate)
		if err != nil {
			if err == ErrNotFound {
				continue
			}
			return Product{}, err
		}

		saved, err := s.repo.Upsert(ctx, fresh)
		if err != nil {
			return Product{}, err
		}
		return saved, nil
	}

	return Product{}, ErrNotFound
}

func (s *Service) FindExactBaseProduct(ctx context.Context, name string) (BaseProduct, error) {
	if s == nil || s.baseProducts == nil {
		return BaseProduct{}, errcode.NotFound
	}
	return s.baseProducts.FindExactByName(ctx, name)
}

func (s *Service) SuggestBaseProducts(ctx context.Context, name string) ([]BaseProduct, error) {
	if s == nil || s.baseProducts == nil {
		return nil, nil
	}

	items, err := s.baseProducts.Search(ctx, name, 10)
	if err != nil {
		return nil, err
	}
	return dedupeBaseProducts(items), nil
}

func (s *Service) CreateCustomProduct(ctx context.Context, req CustomProductRequest) (BaseProduct, error) {
	if strings.TrimSpace(req.Name) == "" ||
		req.Calories < 0 ||
		req.Protein < 0 ||
		req.Fat < 0 ||
		req.Carbs < 0 {
		return BaseProduct{}, errcode.InvalidNutrients
	}

	if s == nil || s.baseProducts == nil {
		return BaseProduct{}, errcode.NotFound
	}

	return s.baseProducts.Create(ctx, BaseProduct{
		Name:     textnorm.LowerTrim(req.Name),
		Calories: req.Calories,
		Protein:  req.Protein,
		Fat:      req.Fat,
		Carbs:    req.Carbs,
	})
}

func dedupeBaseProducts(items []BaseProduct) []BaseProduct {
	if len(items) < 2 {
		return items
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]BaseProduct, 0, len(items))
	for _, item := range items {
		key := textnorm.LowerTrim(item.Name)
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
