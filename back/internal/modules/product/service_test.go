package product

import (
	"context"
	"strings"
	"testing"

	"back/internal/platform/errcode"

	"github.com/stretchr/testify/require"
)

type memoryBarcodeRepo struct {
	product Product
	has     bool
	saved   bool
}

func (r *memoryBarcodeRepo) GetByBarcode(_ context.Context, barcode string) (Product, error) {
	if r.has && r.product.Barcode == barcode {
		return r.product, nil
	}
	return Product{}, ErrNotFound
}

func (r *memoryBarcodeRepo) Upsert(_ context.Context, p Product) (Product, error) {
	r.saved = true
	r.product = p
	r.has = true
	return p, nil
}

type memoryProvider struct {
	product Product
	called  bool
}

func (p *memoryProvider) FetchByBarcode(_ context.Context, barcode string) (Product, error) {
	p.called = true
	if p.product.Barcode == barcode {
		return p.product, nil
	}
	return Product{}, ErrNotFound
}

type memoryBaseProductRepository struct {
	items map[string]BaseProduct
}

func (r *memoryBaseProductRepository) Search(_ context.Context, query string, limit int) ([]BaseProduct, error) {
	if limit <= 0 {
		limit = 10
	}
	out := make([]BaseProduct, 0, limit)
	for _, item := range r.items {
		if strings.Contains(item.Name, query) {
			out = append(out, item)
		}
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (r *memoryBaseProductRepository) FindExactByName(_ context.Context, name string) (BaseProduct, error) {
	if item, ok := r.items[name]; ok {
		return item, nil
	}
	return BaseProduct{}, errcode.NotFound
}

func (r *memoryBaseProductRepository) FindFuzzyByName(_ context.Context, name string) ([]BaseProduct, error) {
	out := make([]BaseProduct, 0, 10)
	for _, item := range r.items {
		if strings.Contains(item.Name, name) {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *memoryBaseProductRepository) Create(_ context.Context, p BaseProduct) (BaseProduct, error) {
	if r.items == nil {
		r.items = map[string]BaseProduct{}
	}
	p.ID = int64(len(r.items) + 1)
	r.items[p.Name] = p
	return p, nil
}

func TestServiceLookupByBarcode(t *testing.T) {
	tests := []struct {
		name            string
		inputBarcode    string
		repo            *memoryBarcodeRepo
		provider        *memoryProvider
		wantName        string
		wantSource      string
		wantProvider    bool
		wantSavedToRepo bool
	}{
		{
			name:         "uses cache when barcode already exists",
			inputBarcode: "12345678",
			repo: &memoryBarcodeRepo{
				has:     true,
				product: Product{Barcode: "12345678", Name: "Cached", Source: "openfoodfacts"},
			},
			provider:        &memoryProvider{product: Product{Barcode: "12345678", Name: "Fresh"}},
			wantName:        "Cached",
			wantSource:      "cache",
			wantProvider:    false,
			wantSavedToRepo: false,
		},
		{
			name:            "fetches and saves when cache misses",
			inputBarcode:    "12345678",
			repo:            &memoryBarcodeRepo{},
			provider:        &memoryProvider{product: Product{Barcode: "12345678", Name: "Fresh", Source: "openfoodfacts"}},
			wantName:        "Fresh",
			wantSource:      "openfoodfacts",
			wantProvider:    true,
			wantSavedToRepo: true,
		},
		{
			name:         "normalizes scanner formatting before lookup",
			inputBarcode: "  4870-0280-02852  ",
			repo: &memoryBarcodeRepo{
				has:     true,
				product: Product{Barcode: "4870028002852", Name: "Normalized", Source: "openfoodfacts"},
			},
			provider:        &memoryProvider{product: Product{Barcode: "4870028002852", Name: "Fresh"}},
			wantName:        "Normalized",
			wantSource:      "cache",
			wantProvider:    false,
			wantSavedToRepo: false,
		},
		{
			name:            "tries UPC and EAN variants",
			inputBarcode:    "123456789012",
			repo:            &memoryBarcodeRepo{},
			provider:        &memoryProvider{product: Product{Barcode: "0123456789012", Name: "UPC Variant", Source: "openfoodfacts"}},
			wantName:        "UPC Variant",
			wantSource:      "openfoodfacts",
			wantProvider:    true,
			wantSavedToRepo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, tt.provider, &memoryBaseProductRepository{})

			got, err := svc.LookupByBarcode(context.Background(), tt.inputBarcode)
			require.NoError(t, err)
			require.Equal(t, tt.wantName, got.Name)
			require.Equal(t, tt.wantSource, got.Source)
			require.Equal(t, tt.wantProvider, tt.provider.called)
			require.Equal(t, tt.wantSavedToRepo, tt.repo.saved)
		})
	}
}

func TestBarcodeLookupCandidates(t *testing.T) {
	require.Equal(t, []string{"4870028002852"}, barcodeLookupCandidates(" 4870-0280-02852 "))
	require.Equal(t, []string{"123456789012", "0123456789012"}, barcodeLookupCandidates("123456789012"))
	require.Equal(t, []string{"0123456789012", "123456789012"}, barcodeLookupCandidates("0123456789012"))
	require.Nil(t, barcodeLookupCandidates("12ab"))
}

func TestServiceCreateCustomProduct(t *testing.T) {
	tests := []struct {
		name       string
		req        CustomProductRequest
		wantErr    error
		wantStored BaseProduct
	}{
		{
			name: "custom product is saved normalized and reusable",
			req: CustomProductRequest{
				Name:     "  Cottage   Cheese  ",
				Calories: 98,
				Protein:  11,
				Fat:      4,
				Carbs:    3.4,
			},
			wantStored: BaseProduct{
				Name:     "cottage cheese",
				Calories: 98,
				Protein:  11,
				Fat:      4,
				Carbs:    3.4,
			},
		},
		{
			name: "zero nutrients are valid",
			req: CustomProductRequest{
				Name:     "water",
				Calories: 0,
				Protein:  0,
				Fat:      0,
				Carbs:    0,
			},
			wantStored: BaseProduct{
				Name:     "water",
				Calories: 0,
				Protein:  0,
				Fat:      0,
				Carbs:    0,
			},
		},
		{
			name: "empty name returns INVALID_NUTRIENTS",
			req: CustomProductRequest{
				Name:     "",
				Calories: 10,
				Protein:  1,
				Fat:      1,
				Carbs:    1,
			},
			wantErr: errcode.InvalidNutrients,
		},
		{
			name: "negative nutrients return INVALID_NUTRIENTS",
			req: CustomProductRequest{
				Name:     "invalid",
				Calories: -1,
				Protein:  0,
				Fat:      0,
				Carbs:    0,
			},
			wantErr: errcode.InvalidNutrients,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseRepo := &memoryBaseProductRepository{}
			svc := NewService(&memoryBarcodeRepo{}, &memoryProvider{}, baseRepo)

			saved, err := svc.CreateCustomProduct(context.Background(), tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantStored.Name, saved.Name)
			require.InDelta(t, tt.wantStored.Calories, saved.Calories, 0.000001)
			require.InDelta(t, tt.wantStored.Protein, saved.Protein, 0.000001)
			require.InDelta(t, tt.wantStored.Fat, saved.Fat, 0.000001)
			require.InDelta(t, tt.wantStored.Carbs, saved.Carbs, 0.000001)

			reloaded, err := svc.FindExactBaseProduct(context.Background(), tt.wantStored.Name)
			require.NoError(t, err)
			require.Equal(t, saved.ID, reloaded.ID)
			require.Equal(t, tt.wantStored.Name, reloaded.Name)
			require.InDelta(t, tt.wantStored.Calories, reloaded.Calories, 0.000001)
			require.InDelta(t, tt.wantStored.Protein, reloaded.Protein, 0.000001)
			require.InDelta(t, tt.wantStored.Fat, reloaded.Fat, 0.000001)
			require.InDelta(t, tt.wantStored.Carbs, reloaded.Carbs, 0.000001)
		})
	}
}
