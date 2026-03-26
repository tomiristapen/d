package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"back/internal/modules/auth"
	"back/internal/modules/diary"
	"back/internal/modules/ingredient"
	"back/internal/modules/input"
	"back/internal/modules/nutrition"
	"back/internal/modules/onboarding"
	"back/internal/modules/product"
	"back/internal/modules/profile"
	"back/internal/modules/recipe"
	"back/internal/modules/user"
	"back/internal/platform/errcode"
	"back/internal/platform/jwtutil"

	"github.com/stretchr/testify/require"
)

type memoryProfileRepo struct{}

func (memoryProfileRepo) Save(_ context.Context, _ profile.Profile) error { return nil }
func (memoryProfileRepo) GetByUserID(_ context.Context, _ string) (profile.Profile, error) {
	return profile.Profile{}, profile.ErrNotFound
}
func (memoryProfileRepo) Delete(_ context.Context, _ string) error { return nil }

type memoryIngredientRepo struct{}

func (memoryIngredientRepo) Search(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (memoryIngredientRepo) FindBestMatch(_ context.Context, _ string) (ingredient.Ingredient, error) {
	return ingredient.Ingredient{}, ingredient.ErrNotFound
}

type memoryProductRepo struct {
	item product.Product
	err  error
}

func (r memoryProductRepo) GetByBarcode(_ context.Context, _ string) (product.Product, error) {
	return r.item, r.err
}

func (r memoryProductRepo) Upsert(_ context.Context, p product.Product) (product.Product, error) {
	return p, nil
}

type timeoutProvider struct{}

func (timeoutProvider) FetchByBarcode(_ context.Context, _ string) (product.Product, error) {
	return product.Product{}, context.DeadlineExceeded
}

type memoryUserRepo struct{}

func (memoryUserRepo) DeleteByID(_ context.Context, _ string) error { return nil }

type memoryDiaryRepo struct{}

func (memoryDiaryRepo) Create(_ context.Context, e diary.Entry) (diary.Entry, error) { return e, nil }
func (memoryDiaryRepo) GetDailyTotals(_ context.Context, userID string, day time.Time) (diary.DailyTotals, error) {
	return diary.DailyTotals{UserID: userID, Date: day.Format("2006-01-02")}, nil
}
func (memoryDiaryRepo) GetTargets(_ context.Context, _ string) (nutrition.Nutrients, error) {
	return nutrition.Nutrients{Calories: 2000, Protein: 120, Fat: 70, Carbs: 240}, nil
}
func (memoryDiaryRepo) DeleteEntry(_ context.Context, _ string, _ int64) error { return nil }
func (memoryDiaryRepo) UpdateEntry(_ context.Context, _ string, _ diary.UpdateEntryInput) (diary.Entry, error) {
	return diary.Entry{}, nil
}

type memoryBaseProductRepo struct {
	items map[string]product.BaseProduct
}

func (r *memoryBaseProductRepo) Search(_ context.Context, query string, limit int) ([]product.BaseProduct, error) {
	keys := make([]string, 0, len(r.items))
	for name := range r.items {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	out := make([]product.BaseProduct, 0, limit)
	for _, name := range keys {
		item := r.items[name]
		if strings.Contains(item.Name, query) {
			out = append(out, item)
		}
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (r *memoryBaseProductRepo) FindExactByName(_ context.Context, name string) (product.BaseProduct, error) {
	if item, ok := r.items[name]; ok {
		return item, nil
	}
	return product.BaseProduct{}, errcode.NotFound
}

func (r *memoryBaseProductRepo) FindFuzzyByName(_ context.Context, name string) ([]product.BaseProduct, error) {
	keys := make([]string, 0, len(r.items))
	for itemName := range r.items {
		keys = append(keys, itemName)
	}
	sort.Strings(keys)

	out := make([]product.BaseProduct, 0, 10)
	for _, itemName := range keys {
		item := r.items[itemName]
		if strings.Contains(item.Name, name) {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *memoryBaseProductRepo) Create(_ context.Context, p product.BaseProduct) (product.BaseProduct, error) {
	if strings.TrimSpace(p.Name) == "" || p.Calories < 0 || p.Protein < 0 || p.Fat < 0 || p.Carbs < 0 {
		return product.BaseProduct{}, errcode.InvalidNutrients
	}
	if r.items == nil {
		r.items = map[string]product.BaseProduct{}
	}
	p.ID = int64(len(r.items) + 1)
	r.items[p.Name] = p
	return p, nil
}

func TestRouterHealthz(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"status"`)
}

func TestRouterSwaggerRoutes(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	t.Run("swagger html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), "swagger-ui")
	})

	t.Run("openapi yaml", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Contains(t, rr.Body.String(), "openapi: 3.0.3")
		require.Contains(t, rr.Body.String(), "/manual/analyze")
	})
}

func TestRouterProducts_UnauthorizedWithoutBearer(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/12345678", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRouterProducts_InvalidBarcode(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/12ab", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRouterProducts_AcceptsNormalizedBarcodeFormatting(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(
		memoryProductRepo{item: product.Product{Barcode: "4870028002852", Name: "Milk"}},
		timeoutProvider{},
		product.NewNoopBaseProductRepository(),
	)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/4870-0280-02852", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"barcode":"4870028002852"`)
}

func TestRouterProducts_UpstreamTimeoutMappedTo504(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	svc := product.NewService(memoryProductRepo{err: product.ErrNotFound}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	manualSvc := input.NewService(svc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(svc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/12345678", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusGatewayTimeout, rr.Code)
}

type captureSaveProfileRepo struct {
	saved bool
}

func (r *captureSaveProfileRepo) Save(_ context.Context, _ profile.Profile) error {
	r.saved = true
	return nil
}
func (r *captureSaveProfileRepo) GetByUserID(_ context.Context, _ string) (profile.Profile, error) {
	return profile.Profile{}, profile.ErrNotFound
}
func (r *captureSaveProfileRepo) Delete(_ context.Context, _ string) error { return nil }

func TestRouterOnboarding_PutProfile_ValidatesActivityLevel(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	repo := &captureSaveProfileRepo{}
	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, product.NewNoopBaseProductRepository())
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(repo)),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/profile", strings.NewReader(`{
  "age": 25,
  "gender": "male",
  "height_cm": 180,
  "weight_kg": 80,
  "activity_level": "",
  "nutrition_goal": "maintain_weight",
  "allergies": [],
  "custom_allergies": [],
  "intolerances": [],
  "diet_type": "none",
  "religious_restriction": "none"
}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.False(t, repo.saved)
}

func TestRouterManualAnalyze_NotFoundReturns404(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{items: map[string]product.BaseProduct{}}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"ghost","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"NOT_FOUND"`)
	require.Contains(t, rr.Body.String(), `"message":"Product not found"`)

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Equal(t, "NOT_FOUND", body.Error.Code)
	require.Equal(t, "Product not found", body.Error.Message)
}

func TestRouterManualAnalyze_SuggestionsHaveConsistentShape(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken breast": {Name: "chicken breast", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
			"chicken thigh":  {Name: "chicken thigh", Calories: 177, Protein: 24, Fat: 8, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"chick","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"product":null`)
	require.Contains(t, rr.Body.String(), `"confidence":null`)
	require.Contains(t, rr.Body.String(), `"suggestions":["chicken breast","chicken thigh"]`)

	var body struct {
		Product     *input.ProductResult `json:"product"`
		Suggestions []string             `json:"suggestions"`
		Confidence  *float64             `json:"confidence"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Nil(t, body.Product)
	require.Nil(t, body.Confidence)
	require.Equal(t, []string{"chicken breast", "chicken thigh"}, body.Suggestions)
}

func TestRouterManualAnalyze_SingleSearchCandidateAutoResolves(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken breast": {Name: "chicken breast", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"breast","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"name":"chicken breast"`)
	require.Contains(t, rr.Body.String(), `"confidence":0.85`)
	require.Contains(t, rr.Body.String(), `"suggestions":[]`)
}

func TestRouterRecipeAnalyze_ReturnsConfidence(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"tomato": {Name: "tomato", Calories: 18, Protein: 0.9, Fat: 0.2, Carbs: 3.9},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipe/analyze", strings.NewReader(`{"name":"salad","ingredients":[{"name":"tomato","amount":100}]}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"confidence":1`)
}

func TestRouterManualCustom_CreatesProduct(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/custom", strings.NewReader(`{"name":"my cake","calories":300,"protein":10,"fat":15,"carbs":40}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Contains(t, rr.Body.String(), `"status":"created"`)
	_, ok := baseRepo.items["my cake"]
	require.True(t, ok)
}

func TestRouterManualAnalyze_ExactMatchNormalizesWhitespace(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken breast": {Name: "chicken breast", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"  CHICKEN   BREAST  ","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"name":"chicken breast"`)
	require.Contains(t, rr.Body.String(), `"calories":165`)
	require.Contains(t, rr.Body.String(), `"protein":31`)
	require.Contains(t, rr.Body.String(), `"confidence":1`)
	require.Contains(t, rr.Body.String(), `"suggestions":[]`)
}

func TestRouterManualAnalyze_InvalidJSONReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"chicken"`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"BAD_REQUEST"`)
	require.Contains(t, rr.Body.String(), `"message":"invalid json body"`)
}

func TestRouterManualAnalyze_MissingAmountReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken breast": {Name: "chicken breast", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"chicken breast"}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"INVALID_AMOUNT"`)
	require.Contains(t, rr.Body.String(), `"message":"Invalid amount"`)
}

func TestRouterManualAnalyze_AmountAboveMaxReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"milk": {Name: "milk", Calories: 42, Protein: 3.4, Fat: 1, Carbs: 5},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"milk","amount":2500}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"INVALID_AMOUNT"`)
}

func TestRouterManualAnalyze_EmptyObjectReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"INVALID_NAME"`)
}

func TestRouterManualAnalyze_MissingContentTypeStillParsesJSON(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"milk": {Name: "milk", Calories: 42, Protein: 3.4, Fat: 1, Carbs: 5},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"milk","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"name":"milk"`)
}

func TestRouterManualAnalyze_InvalidJSONTokenReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{ name: chicken }`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"BAD_REQUEST"`)
	require.Contains(t, rr.Body.String(), `"message":"invalid json body"`)
}

func TestRouterManualAddToDiary_SuggestionsRequireExplicitSelection(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken breast": {Name: "chicken breast", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
			"chicken thigh":  {Name: "chicken thigh", Calories: 177, Protein: 24, Fat: 8, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/manual/add-to-diary", strings.NewReader(`{"name":"chick","amount":100}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"NOT_FOUND"`)
	require.Contains(t, rr.Body.String(), `"message":"Product not found"`)
}

func TestRouterManualCustomThenAnalyze_ReusesCreatedProduct(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/manual/custom", strings.NewReader(`{"name":"  my   cake  ","calories":300,"protein":10,"fat":15,"carbs":40}`))
	createReq.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	createRR := httptest.NewRecorder()
	r.ServeHTTP(createRR, createReq)

	require.Equal(t, http.StatusCreated, createRR.Code)
	_, ok := baseRepo.items["my cake"]
	require.True(t, ok)

	analyzeReq := httptest.NewRequest(http.MethodPost, "/api/v1/manual/analyze", strings.NewReader(`{"name":"MY CAKE","amount":100}`))
	analyzeReq.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	analyzeRR := httptest.NewRecorder()
	r.ServeHTTP(analyzeRR, analyzeReq)

	require.Equal(t, http.StatusOK, analyzeRR.Code)
	require.Contains(t, analyzeRR.Body.String(), `"name":"my cake"`)
	require.Contains(t, analyzeRR.Body.String(), `"calories":300`)
	require.Contains(t, analyzeRR.Body.String(), `"protein":10`)
	require.Contains(t, analyzeRR.Body.String(), `"fat":15`)
	require.Contains(t, analyzeRR.Body.String(), `"carbs":40`)
}

func TestRouterRecipeAnalyze_InvalidJSONReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipe/analyze", strings.NewReader(`{"name":"meal","ingredients":[`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"BAD_REQUEST"`)
	require.Contains(t, rr.Body.String(), `"message":"invalid json body"`)
}

func TestRouterRecipeAnalyze_UsesCustomProduct(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/manual/custom", strings.NewReader(`{"name":"  my   cake  ","calories":300,"protein":10,"fat":15,"carbs":40}`))
	createReq.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	createRR := httptest.NewRecorder()
	r.ServeHTTP(createRR, createReq)

	require.Equal(t, http.StatusCreated, createRR.Code)

	recipeReq := httptest.NewRequest(http.MethodPost, "/api/v1/recipe/analyze", strings.NewReader(`{"name":"dessert","ingredients":[{"name":"  MY   CAKE  ","amount":100}]}`))
	recipeReq.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	recipeRR := httptest.NewRecorder()
	r.ServeHTTP(recipeRR, recipeReq)

	require.Equal(t, http.StatusOK, recipeRR.Code)
	require.Contains(t, recipeRR.Body.String(), `"name":"dessert"`)
	require.Contains(t, recipeRR.Body.String(), `"calories":300`)
	require.Contains(t, recipeRR.Body.String(), `"protein":10`)
	require.Contains(t, recipeRR.Body.String(), `"fat":15`)
	require.Contains(t, recipeRR.Body.String(), `"carbs":40`)
	require.Contains(t, recipeRR.Body.String(), `"ingredients":[{"name":"my cake","amount":100`)
}

func TestRouterRecipeAnalyze_MixedValidAndInvalidReturns404(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"chicken": {Name: "chicken", Calories: 165, Protein: 31, Fat: 3.6, Carbs: 0},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipe/analyze", strings.NewReader(`{"name":"mixed meal","ingredients":[{"name":"chicken","amount":100},{"name":"unknown123","amount":100}]}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"INGREDIENT_NOT_FOUND"`)
}

func TestRouterRecipeAnalyze_IngredientAmountAboveMaxReturns400(t *testing.T) {
	tokens := jwtutil.NewManager("test", "a", "b", time.Minute, time.Hour)
	pair, err := tokens.GeneratePair("user-1")
	require.NoError(t, err)

	ingSvc := ingredient.NewService(memoryIngredientRepo{})
	baseRepo := &memoryBaseProductRepo{
		items: map[string]product.BaseProduct{
			"tomato": {Name: "tomato", Calories: 18, Protein: 0.9, Fat: 0.2, Carbs: 3.9},
		},
	}
	productSvc := product.NewService(memoryProductRepo{}, timeoutProvider{}, baseRepo)
	manualSvc := input.NewService(productSvc)
	recipeSvc := recipe.NewService(manualSvc)
	r := NewRouter(
		auth.NewHandler(new(auth.Service)),
		user.NewHandler(user.NewService(memoryUserRepo{})),
		profile.NewHandler(profile.NewService(memoryProfileRepo{})),
		ingredient.NewHandler(ingSvc),
		product.NewHandler(productSvc),
		input.NewHandler(manualSvc),
		recipe.NewHandler(recipeSvc),
		diary.NewHandler(diary.NewService(memoryDiaryRepo{}), manualSvc, recipeSvc),
		onboarding.NewHandler(onboarding.NewService(onboarding.NewPostgresRepository(nil))),
		tokens,
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/recipe/analyze", strings.NewReader(`{"name":"large meal","ingredients":[{"name":"tomato","amount":2500}]}`))
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"INVALID_AMOUNT"`)
}
