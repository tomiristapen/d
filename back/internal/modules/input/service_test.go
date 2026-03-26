package input

import (
	"context"
	"strings"
	"testing"

	"back/internal/modules/product"
	"back/internal/platform/errcode"

	"github.com/stretchr/testify/require"
)

type stubProductService struct {
	exact             map[string]product.BaseProduct
	suggestions       map[string][]product.BaseProduct
	exactQueries      []string
	suggestionQueries []string
	created           []product.CustomProductRequest
}

func (s *stubProductService) FindExactBaseProduct(_ context.Context, name string) (product.BaseProduct, error) {
	s.exactQueries = append(s.exactQueries, name)
	if item, ok := s.exact[name]; ok {
		return item, nil
	}
	return product.BaseProduct{}, errcode.NotFound
}

func (s *stubProductService) SuggestBaseProducts(_ context.Context, name string) ([]product.BaseProduct, error) {
	s.suggestionQueries = append(s.suggestionQueries, name)
	return append([]product.BaseProduct(nil), s.suggestions[name]...), nil
}

func (s *stubProductService) CreateCustomProduct(_ context.Context, req product.CustomProductRequest) (product.BaseProduct, error) {
	s.created = append(s.created, req)
	name := NormalizeName(req.Name)
	if name == "" || req.Calories < 0 || req.Protein < 0 || req.Fat < 0 || req.Carbs < 0 {
		return product.BaseProduct{}, errcode.InvalidNutrients
	}
	item := product.BaseProduct{
		ID:       int64(len(s.created)),
		Name:     name,
		Calories: req.Calories,
		Protein:  req.Protein,
		Fat:      req.Fat,
		Carbs:    req.Carbs,
	}
	if s.exact == nil {
		s.exact = map[string]product.BaseProduct{}
	}
	s.exact[item.Name] = item
	return item, nil
}

func TestManualAnalyze(t *testing.T) {
	chicken := product.BaseProduct{
		Name:     "chicken",
		Calories: 165,
		Protein:  31,
		Fat:      3.6,
		Carbs:    0,
	}
	chickenBreast := product.BaseProduct{
		Name:     "chicken breast",
		Calories: 165,
		Protein:  31,
		Fat:      3.6,
		Carbs:    0,
	}
	oats := product.BaseProduct{
		Name:     "oats",
		Calories: 389,
		Protein:  16.9,
		Fat:      6.9,
		Carbs:    66.3,
	}
	milk := product.BaseProduct{
		Name:     "milk",
		Calories: 42,
		Protein:  3.4,
		Fat:      1,
		Carbs:    5,
	}

	tests := []struct {
		name               string
		req                ManualAnalyzeRequest
		exact              map[string]product.BaseProduct
		suggestions        map[string][]product.BaseProduct
		wantErr            error
		wantProduct        *ProductResult
		wantSuggestions    []string
		wantExactQueries   []string
		wantSuggestQueries []string
	}{
		{
			name: "exact match returns per 100g values",
			req: ManualAnalyzeRequest{
				Name:   "chicken breast",
				Amount: 100,
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
			},
			wantProduct: &ProductResult{
				Name:     "chicken breast",
				Calories: 165,
				Protein:  31,
				Fat:      3.6,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken breast"},
		},
		{
			name: "exact match normalizes trimmed name",
			req: ManualAnalyzeRequest{
				Name:   "  chicken  ",
				Amount: 200,
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
			},
			wantProduct: &ProductResult{
				Name:     "chicken",
				Calories: 330,
				Protein:  62,
				Fat:      7.2,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken"},
		},
		{
			name: "exact match normalizes uppercase name",
			req: ManualAnalyzeRequest{
				Name:   "CHICKEN",
				Amount: 100,
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
			},
			wantProduct: &ProductResult{
				Name:     "chicken",
				Calories: 165,
				Protein:  31,
				Fat:      3.6,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken"},
		},
		{
			name: "multiple spaces inside name are normalized",
			req: ManualAnalyzeRequest{
				Name:   "chicken   breast",
				Amount: 100,
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
			},
			wantProduct: &ProductResult{
				Name:     "chicken breast",
				Calories: 165,
				Protein:  31,
				Fat:      3.6,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken breast"},
		},
		{
			name: "fuzzy search returns suggestions without auto selection",
			req: ManualAnalyzeRequest{
				Name:   "chick",
				Amount: 100,
			},
			suggestions: map[string][]product.BaseProduct{
				"chick": {
					{Name: "chicken breast"},
					{Name: "chicken thigh"},
				},
			},
			wantSuggestions:    []string{"chicken breast", "chicken thigh"},
			wantExactQueries:   []string{"chick"},
			wantSuggestQueries: []string{"chick"},
		},
		{
			name: "single search candidate auto resolves with lower confidence",
			req: ManualAnalyzeRequest{
				Name:   "breast",
				Amount: 100,
			},
			suggestions: map[string][]product.BaseProduct{
				"breast": {
					chickenBreast,
				},
			},
			wantProduct: &ProductResult{
				Name:     "chicken breast",
				Calories: 165,
				Protein:  31,
				Fat:      3.6,
				Carbs:    0,
			},
			wantExactQueries:   []string{"breast"},
			wantSuggestQueries: []string{"breast"},
		},
		{
			name: "duplicate search candidates are deduped before auto resolve",
			req: ManualAnalyzeRequest{
				Name:   "broccoli",
				Amount: 100,
			},
			suggestions: map[string][]product.BaseProduct{
				"broccoli": {
					{Name: "broccoli, raw", Calories: 31.5, Protein: 2.57, Fat: 0.34, Carbs: 6.29},
					{Name: "broccoli, raw", Calories: 31.5, Protein: 2.57, Fat: 0.34, Carbs: 6.29},
				},
			},
			wantProduct: &ProductResult{
				Name:     "broccoli, raw",
				Calories: 31.5,
				Protein:  2.57,
				Fat:      0.34,
				Carbs:    6.29,
			},
			wantExactQueries:   []string{"broccoli"},
			wantSuggestQueries: []string{"broccoli"},
		},
		{
			name: "selecting suggestion by exact name returns scaled product",
			req: ManualAnalyzeRequest{
				Name:   "chicken breast",
				Amount: 150,
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
			},
			wantProduct: &ProductResult{
				Name:     "chicken breast",
				Calories: 247.5,
				Protein:  46.5,
				Fat:      5.4,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken breast"},
		},
		{
			name: "fuzzy empty list returns NOT_FOUND",
			req: ManualAnalyzeRequest{
				Name:   "zzzxxyyqwe",
				Amount: 100,
			},
			wantErr:            errcode.NotFound,
			wantExactQueries:   []string{"zzzxxyyqwe"},
			wantSuggestQueries: []string{"zzzxxyyqwe"},
		},
		{
			name: "small amount scales correctly",
			req: ManualAnalyzeRequest{
				Name:   "chicken breast",
				Amount: 1,
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
			},
			wantProduct: &ProductResult{
				Name:     "chicken breast",
				Calories: 1.65,
				Protein:  0.31,
				Fat:      0.036,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken breast"},
		},
		{
			name: "large in-range amount scales correctly",
			req: ManualAnalyzeRequest{
				Name:   "oats",
				Amount: 1500,
			},
			exact: map[string]product.BaseProduct{
				"oats": oats,
			},
			wantProduct: &ProductResult{
				Name:     "oats",
				Calories: 5835,
				Protein:  253.5,
				Fat:      103.5,
				Carbs:    994.5,
			},
			wantExactQueries: []string{"oats"},
		},
		{
			name: "amount 2000 is valid",
			req: ManualAnalyzeRequest{
				Name:   "chicken",
				Amount: 2000,
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
			},
			wantProduct: &ProductResult{
				Name:     "chicken",
				Calories: 3300,
				Protein:  620,
				Fat:      72,
				Carbs:    0,
			},
			wantExactQueries: []string{"chicken"},
		},
		{
			name: "boundary amount 0.0001 is valid",
			req: ManualAnalyzeRequest{
				Name:   "milk",
				Amount: 0.0001,
			},
			exact: map[string]product.BaseProduct{
				"milk": milk,
			},
			wantProduct: &ProductResult{
				Name:     "milk",
				Calories: 0.000042,
				Protein:  0.0000034,
				Fat:      0.000001,
				Carbs:    0.000005,
			},
			wantExactQueries: []string{"milk"},
		},
		{
			name: "boundary amount 100 keeps per 100g values",
			req: ManualAnalyzeRequest{
				Name:   "milk",
				Amount: 100,
			},
			exact: map[string]product.BaseProduct{
				"milk": milk,
			},
			wantProduct: &ProductResult{
				Name:     "milk",
				Calories: 42,
				Protein:  3.4,
				Fat:      1,
				Carbs:    5,
			},
			wantExactQueries: []string{"milk"},
		},
		{
			name: "multiple fuzzy matches returns full suggestion list",
			req: ManualAnalyzeRequest{
				Name:   "berry",
				Amount: 100,
			},
			suggestions: map[string][]product.BaseProduct{
				"berry": {
					{Name: "blueberry"},
					{Name: "blackberry"},
					{Name: "strawberry"},
				},
			},
			wantSuggestions:    []string{"blueberry", "blackberry", "strawberry"},
			wantExactQueries:   []string{"berry"},
			wantSuggestQueries: []string{"berry"},
		},
		{
			name: "empty name returns INVALID_NAME",
			req: ManualAnalyzeRequest{
				Name:   "",
				Amount: 100,
			},
			wantErr: errcode.InvalidName,
		},
		{
			name: "amount less than or equal to zero returns INVALID_AMOUNT",
			req: ManualAnalyzeRequest{
				Name:   "milk",
				Amount: 0,
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "negative amount returns INVALID_AMOUNT",
			req: ManualAnalyzeRequest{
				Name:   "milk",
				Amount: -50,
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "amount above max returns INVALID_AMOUNT",
			req: ManualAnalyzeRequest{
				Name:   "milk",
				Amount: 2500,
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "very long name returns controlled NOT_FOUND",
			req: ManualAnalyzeRequest{
				Name:   strings.Repeat("chicken", 150),
				Amount: 100,
			},
			wantErr:            errcode.NotFound,
			wantExactQueries:   []string{strings.Repeat("chicken", 150)},
			wantSuggestQueries: []string{strings.Repeat("chicken", 150)},
		},
		{
			name: "special characters do not crash and return NOT_FOUND",
			req: ManualAnalyzeRequest{
				Name:   "chicken!!!@@@###",
				Amount: 100,
			},
			wantErr:            errcode.NotFound,
			wantExactQueries:   []string{"chicken!!!@@@###"},
			wantSuggestQueries: []string{"chicken!!!@@@###"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products := &stubProductService{
				exact:       tt.exact,
				suggestions: tt.suggestions,
			}
			svc := NewService(products)

			got, err := svc.Analyze(context.Background(), tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.wantExactQueries, products.exactQueries)
				require.Equal(t, tt.wantSuggestQueries, products.suggestionQueries)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantExactQueries, products.exactQueries)
			require.Equal(t, tt.wantSuggestQueries, products.suggestionQueries)

			if tt.wantProduct != nil {
				require.NotNil(t, got.Product)
				require.NotNil(t, got.Confidence)
				require.Equal(t, []string{}, got.Suggestions)
				if len(tt.wantSuggestQueries) > 0 {
					require.InDelta(t, singleMatchConfidence, *got.Confidence, 0.000001)
				} else {
					require.InDelta(t, 1.0, *got.Confidence, 0.000001)
				}
				require.InDelta(t, tt.wantProduct.Calories, got.Product.Calories, 0.000001)
				require.InDelta(t, tt.wantProduct.Protein, got.Product.Protein, 0.000001)
				require.InDelta(t, tt.wantProduct.Fat, got.Product.Fat, 0.000001)
				require.InDelta(t, tt.wantProduct.Carbs, got.Product.Carbs, 0.000001)
				require.Equal(t, tt.wantProduct.Name, got.Product.Name)
				require.InDelta(t, tt.req.Amount, got.AmountG, 0.000001)
				return
			}

			require.Nil(t, got.Product)
			require.Nil(t, got.Confidence)
			require.Equal(t, tt.wantSuggestions, got.Suggestions)
			require.InDelta(t, tt.req.Amount, got.AmountG, 0.000001)
		})
	}
}

func TestCreateCustomProductThenAnalyze(t *testing.T) {
	products := &stubProductService{}
	svc := NewService(products)

	err := svc.CreateCustomProduct(context.Background(), CustomProductRequest{
		Name:     "  My   Cake  ",
		Calories: 300,
		Protein:  10,
		Fat:      15,
		Carbs:    40,
	})
	require.NoError(t, err)

	got, err := svc.Analyze(context.Background(), ManualAnalyzeRequest{
		Name:   "MY CAKE",
		Amount: 100,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"my cake"}, products.exactQueries)
	require.NotNil(t, got.Product)
	require.NotNil(t, got.Confidence)
	require.Equal(t, []string{}, got.Suggestions)
	require.Equal(t, "my cake", got.Product.Name)
	require.InDelta(t, 300, got.Product.Calories, 0.000001)
	require.InDelta(t, 10, got.Product.Protein, 0.000001)
	require.InDelta(t, 15, got.Product.Fat, 0.000001)
	require.InDelta(t, 40, got.Product.Carbs, 0.000001)
	require.InDelta(t, 1.0, *got.Confidence, 0.000001)
}

func TestCreateCustomProduct(t *testing.T) {
	tests := []struct {
		name       string
		req        CustomProductRequest
		wantErr    error
		wantCreate product.CustomProductRequest
	}{
		{
			name: "valid custom product is forwarded",
			req: CustomProductRequest{
				Name:     "my cake",
				Calories: 300,
				Protein:  10,
				Fat:      15,
				Carbs:    40,
			},
			wantCreate: product.CustomProductRequest{
				Name:     "my cake",
				Calories: 300,
				Protein:  10,
				Fat:      15,
				Carbs:    40,
			},
		},
		{
			name: "invalid nutrients bubble up",
			req: CustomProductRequest{
				Name:     "",
				Calories: 300,
			},
			wantErr: errcode.InvalidNutrients,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products := &stubProductService{}
			svc := NewService(products)

			err := svc.CreateCustomProduct(context.Background(), tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Len(t, products.created, 1)
			require.Equal(t, tt.wantCreate, products.created[0])
		})
	}
}
