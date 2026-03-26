package recipe

import (
	"context"
	"strings"
	"testing"

	"back/internal/modules/input"
	"back/internal/modules/product"
	"back/internal/platform/errcode"

	"github.com/stretchr/testify/require"
)

type stubProductService struct {
	exact             map[string]product.BaseProduct
	suggestions       map[string][]product.BaseProduct
	exactQueries      []string
	suggestionQueries []string
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
	name := input.NormalizeName(req.Name)
	if name == "" || req.Calories < 0 || req.Protein < 0 || req.Fat < 0 || req.Carbs < 0 {
		return product.BaseProduct{}, errcode.InvalidNutrients
	}
	item := product.BaseProduct{
		ID:       int64(len(s.exact) + 1),
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

func TestRecipeAnalyze(t *testing.T) {
	chicken := product.BaseProduct{
		Name:     "chicken",
		Calories: 165,
		Protein:  31,
		Fat:      3.6,
		Carbs:    0,
	}
	tomato := product.BaseProduct{
		Name:     "tomato",
		Calories: 18,
		Protein:  0.9,
		Fat:      0.2,
		Carbs:    3.9,
	}
	chickenBreast := product.BaseProduct{
		Name:     "chicken breast",
		Calories: 165,
		Protein:  31,
		Fat:      3.6,
		Carbs:    0,
	}
	rice := product.BaseProduct{
		Name:     "rice",
		Calories: 130,
		Protein:  2.7,
		Fat:      0.3,
		Carbs:    28,
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
		req                AnalyzeRequest
		exact              map[string]product.BaseProduct
		suggestions        map[string][]product.BaseProduct
		wantErr            error
		wantProduct        input.ProductResult
		wantIngredients    []ResolvedIngredient
		wantExactQueries   []string
		wantSuggestQueries []string
	}{
		{
			name: "multiple ingredients aggregate correctly",
			req: AnalyzeRequest{
				Name: "meal",
				Ingredients: []IngredientInput{
					{Name: "chicken breast", Amount: 200},
					{Name: "rice", Amount: 100},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
				"rice":           rice,
			},
			wantProduct: input.ProductResult{
				Name:     "meal",
				Calories: 460,
				Protein:  64.7,
				Fat:      7.5,
				Carbs:    28,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "chicken breast", Amount: 200, Calories: 330, Protein: 62, Fat: 7.2, Carbs: 0},
				{Name: "rice", Amount: 100, Calories: 130, Protein: 2.7, Fat: 0.3, Carbs: 28},
			},
			wantExactQueries: []string{"chicken breast", "rice"},
		},
		{
			name: "recipe aggregates scaled nutrients",
			req: AnalyzeRequest{
				Name: "chicken salad",
				Ingredients: []IngredientInput{
					{Name: "chicken", Amount: 200},
					{Name: "tomato", Amount: 100},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
				"tomato":  tomato,
			},
			wantProduct: input.ProductResult{
				Name:     "chicken salad",
				Calories: 348,
				Protein:  62.9,
				Fat:      7.4,
				Carbs:    3.9,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "chicken", Amount: 200, Calories: 330, Protein: 62, Fat: 7.2, Carbs: 0},
				{Name: "tomato", Amount: 100, Calories: 18, Protein: 0.9, Fat: 0.2, Carbs: 3.9},
			},
			wantExactQueries: []string{"chicken", "tomato"},
		},
		{
			name: "duplicate ingredients are merged after normalization with inner spaces",
			req: AnalyzeRequest{
				Name: "double chicken",
				Ingredients: []IngredientInput{
					{Name: "  chicken   breast  ", Amount: 100},
					{Name: "CHICKEN BREAST", Amount: 200},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken breast": chickenBreast,
			},
			wantProduct: input.ProductResult{
				Name:     "double chicken",
				Calories: 495,
				Protein:  93,
				Fat:      10.8,
				Carbs:    0,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "chicken breast", Amount: 300, Calories: 495, Protein: 93, Fat: 10.8, Carbs: 0},
			},
			wantExactQueries: []string{"chicken breast"},
		},
		{
			name: "duplicate ingredients are merged after normalization",
			req: AnalyzeRequest{
				Name: "",
				Ingredients: []IngredientInput{
					{Name: "  chicken  ", Amount: 50},
					{Name: "CHICKEN", Amount: 150},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
			},
			wantProduct: input.ProductResult{
				Name:     "",
				Calories: 330,
				Protein:  62,
				Fat:      7.2,
				Carbs:    0,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "chicken", Amount: 200, Calories: 330, Protein: 62, Fat: 7.2, Carbs: 0},
			},
			wantExactQueries: []string{"chicken"},
		},
		{
			name: "very small amount ingredient is valid",
			req: AnalyzeRequest{
				Name: "tiny milk",
				Ingredients: []IngredientInput{
					{Name: "milk", Amount: 0.0001},
				},
			},
			exact: map[string]product.BaseProduct{
				"milk": milk,
			},
			wantProduct: input.ProductResult{
				Name:     "tiny milk",
				Calories: 0.000042,
				Protein:  0.0000034,
				Fat:      0.000001,
				Carbs:    0.000005,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "milk", Amount: 0.0001, Calories: 0.000042, Protein: 0.0000034, Fat: 0.000001, Carbs: 0.000005},
			},
			wantExactQueries: []string{"milk"},
		},
		{
			name: "large in-range amounts aggregate correctly",
			req: AnalyzeRequest{
				Name: "bulk meal",
				Ingredients: []IngredientInput{
					{Name: "chicken", Amount: 1500},
					{Name: "tomato", Amount: 500},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
				"tomato":  tomato,
			},
			wantProduct: input.ProductResult{
				Name:     "bulk meal",
				Calories: 2565,
				Protein:  469.5,
				Fat:      55,
				Carbs:    19.5,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "chicken", Amount: 1500, Calories: 2475, Protein: 465, Fat: 54, Carbs: 0},
				{Name: "tomato", Amount: 500, Calories: 90, Protein: 4.5, Fat: 1, Carbs: 19.5},
			},
			wantExactQueries: []string{"chicken", "tomato"},
		},
		{
			name: "single ingredient recipe is valid",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "tomato", Amount: 100},
				},
			},
			exact: map[string]product.BaseProduct{
				"tomato": tomato,
			},
			wantProduct: input.ProductResult{
				Name:     "",
				Calories: 18,
				Protein:  0.9,
				Fat:      0.2,
				Carbs:    3.9,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "tomato", Amount: 100, Calories: 18, Protein: 0.9, Fat: 0.2, Carbs: 3.9},
			},
			wantExactQueries: []string{"tomato"},
		},
		{
			name: "mixed valid and invalid ingredients return INGREDIENT_NOT_FOUND",
			req: AnalyzeRequest{
				Name: "mixed meal",
				Ingredients: []IngredientInput{
					{Name: "chicken", Amount: 100},
					{Name: "unknown123", Amount: 100},
				},
			},
			exact: map[string]product.BaseProduct{
				"chicken": chicken,
			},
			wantErr:            errcode.IngredientNotFound,
			wantExactQueries:   []string{"chicken", "unknown123"},
			wantSuggestQueries: []string{"unknown123"},
		},
		{
			name:    "empty recipe returns EMPTY_RECIPE",
			req:     AnalyzeRequest{},
			wantErr: errcode.EmptyRecipe,
		},
		{
			name: "empty ingredient name returns INVALID_NAME",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "", Amount: 100},
				},
			},
			wantErr: errcode.InvalidName,
		},
		{
			name: "invalid ingredient amount returns INVALID_AMOUNT",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "tomato", Amount: 0},
				},
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "negative ingredient amount returns INVALID_AMOUNT",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "tomato", Amount: -50},
				},
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "ingredient amount above max returns INVALID_AMOUNT",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "tomato", Amount: 2500},
				},
			},
			wantErr: errcode.InvalidAmount,
		},
		{
			name: "ingredient with single search candidate auto resolves",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "mystery", Amount: 100},
				},
			},
			suggestions: map[string][]product.BaseProduct{
				"mystery": {
					{Name: "mystery soup", Calories: 80, Protein: 4, Fat: 2, Carbs: 10},
				},
			},
			wantProduct: input.ProductResult{
				Name:     "",
				Calories: 80,
				Protein:  4,
				Fat:      2,
				Carbs:    10,
			},
			wantIngredients: []ResolvedIngredient{
				{Name: "mystery soup", Amount: 100, Calories: 80, Protein: 4, Fat: 2, Carbs: 10},
			},
			wantExactQueries:   []string{"mystery"},
			wantSuggestQueries: []string{"mystery"},
		},
		{
			name: "ingredient with multiple suggestions returns INGREDIENT_NOT_FOUND",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "mystery", Amount: 100},
				},
			},
			suggestions: map[string][]product.BaseProduct{
				"mystery": {
					{Name: "mystery soup"},
					{Name: "mystery stew"},
				},
			},
			wantErr:            errcode.IngredientNotFound,
			wantExactQueries:   []string{"mystery"},
			wantSuggestQueries: []string{"mystery"},
		},
		{
			name: "ingredient not found without suggestions returns INGREDIENT_NOT_FOUND",
			req: AnalyzeRequest{
				Ingredients: []IngredientInput{
					{Name: "ghost pepper candy", Amount: 100},
				},
			},
			wantErr:            errcode.IngredientNotFound,
			wantExactQueries:   []string{"ghost pepper candy"},
			wantSuggestQueries: []string{"ghost pepper candy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products := &stubProductService{
				exact:       tt.exact,
				suggestions: tt.suggestions,
			}
			manual := input.NewService(products)
			svc := NewService(manual)

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

			require.Equal(t, tt.wantProduct.Name, got.Product.Name)
			require.InDelta(t, tt.wantProduct.Calories, got.Product.Calories, 0.000001)
			require.InDelta(t, tt.wantProduct.Protein, got.Product.Protein, 0.000001)
			require.InDelta(t, tt.wantProduct.Fat, got.Product.Fat, 0.000001)
			require.InDelta(t, tt.wantProduct.Carbs, got.Product.Carbs, 0.000001)
			require.InDelta(t, 1.0, got.Confidence, 0.000001)
			require.Len(t, got.Ingredients, len(tt.wantIngredients))

			var totalAmount float64
			for i := range tt.wantIngredients {
				totalAmount += tt.wantIngredients[i].Amount
				require.Equal(t, tt.wantIngredients[i].Name, got.Ingredients[i].Name)
				require.InDelta(t, tt.wantIngredients[i].Amount, got.Ingredients[i].Amount, 0.000001)
				require.InDelta(t, tt.wantIngredients[i].Calories, got.Ingredients[i].Calories, 0.000001)
				require.InDelta(t, tt.wantIngredients[i].Protein, got.Ingredients[i].Protein, 0.000001)
				require.InDelta(t, tt.wantIngredients[i].Fat, got.Ingredients[i].Fat, 0.000001)
				require.InDelta(t, tt.wantIngredients[i].Carbs, got.Ingredients[i].Carbs, 0.000001)
			}
			require.InDelta(t, totalAmount, got.AmountG, 0.000001)
		})
	}
}

func TestRecipeAnalyze_UsesCustomProduct(t *testing.T) {
	products := &stubProductService{}
	manual := input.NewService(products)
	err := manual.CreateCustomProduct(context.Background(), input.CustomProductRequest{
		Name:     "  My   Cake  ",
		Calories: 300,
		Protein:  10,
		Fat:      15,
		Carbs:    40,
	})
	require.NoError(t, err)

	svc := NewService(manual)
	got, err := svc.Analyze(context.Background(), AnalyzeRequest{
		Name: "dessert",
		Ingredients: []IngredientInput{
			{Name: "MY CAKE", Amount: 100},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"my cake"}, products.exactQueries)
	require.Equal(t, "dessert", got.Product.Name)
	require.InDelta(t, 300, got.Product.Calories, 0.000001)
	require.InDelta(t, 10, got.Product.Protein, 0.000001)
	require.InDelta(t, 15, got.Product.Fat, 0.000001)
	require.InDelta(t, 40, got.Product.Carbs, 0.000001)
	require.Len(t, got.Ingredients, 1)
	require.Equal(t, "my cake", got.Ingredients[0].Name)
}

func TestRecipeAnalyze_ManyIngredients(t *testing.T) {
	exact := make(map[string]product.BaseProduct, 24)
	ingredients := make([]IngredientInput, 0, 24)

	for i := 1; i <= 24; i++ {
		name := input.NormalizeName("item " + string(rune('a'+i-1)))
		exact[name] = product.BaseProduct{
			Name:     name,
			Calories: float64(i * 10),
			Protein:  float64(i),
			Fat:      float64(i) / 10,
			Carbs:    float64(i * 2),
		}
		ingredients = append(ingredients, IngredientInput{
			Name:   strings.ToUpper(name),
			Amount: 100,
		})
	}

	products := &stubProductService{exact: exact}
	manual := input.NewService(products)
	svc := NewService(manual)

	got, err := svc.Analyze(context.Background(), AnalyzeRequest{
		Name:        "big recipe",
		Ingredients: ingredients,
	})
	require.NoError(t, err)
	require.Len(t, got.Ingredients, 24)
	require.Len(t, products.exactQueries, 24)
	require.InDelta(t, 3000, got.Product.Calories, 0.000001)
	require.InDelta(t, 300, got.Product.Protein, 0.000001)
	require.InDelta(t, 30, got.Product.Fat, 0.000001)
	require.InDelta(t, 600, got.Product.Carbs, 0.000001)
	require.InDelta(t, 2400, got.AmountG, 0.000001)
}
