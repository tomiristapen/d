package httpapi

import (
	"net/http"
	"os"
	"strings"

	"back/internal/modules/auth"
	"back/internal/modules/diary"
	"back/internal/modules/ingredient"
	"back/internal/modules/input"
	"back/internal/modules/onboarding"
	"back/internal/modules/product"
	"back/internal/modules/profile"
	"back/internal/modules/recipe"
	"back/internal/modules/user"
	"back/internal/platform/httpx"
	"back/internal/platform/jwtutil"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	authHandler *auth.Handler,
	userHandler *user.Handler,
	profileHandler *profile.Handler,
	ingredientHandler *ingredient.Handler,
	productHandler *product.Handler,
	manualHandler *input.Handler,
	recipeHandler *recipe.Handler,
	diaryHandler *diary.Handler,
	onboardingHandler *onboarding.Handler,
	tokenManager *jwtutil.Manager,
) http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware())
	r.Get("/swagger", swaggerUIHandler())
	r.Get("/swagger/", swaggerUIHandler())
	r.Get("/swagger/openapi.yaml", swaggerSpecHandler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/onboarding/options", onboardingHandler.GetV1Options)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.PostRegister)
			r.Post("/send-verification-code", authHandler.PostSendVerificationCode)
			r.Post("/verify-email", authHandler.PostVerifyEmail)
			r.Post("/login", authHandler.PostLogin)
			r.Post("/send-login-code", authHandler.PostSendLoginCode)
			r.Post("/send-password-reset-code", authHandler.PostSendPasswordResetCode)
			r.Post("/login-with-code", authHandler.PostLoginWithCode)
			r.Post("/reset-password", authHandler.PostResetPassword)
			r.Post("/google", authHandler.PostGoogle)
			r.Post("/refresh", authHandler.PostRefresh)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware(tokenManager))

			r.Post("/auth/set-password", authHandler.PostSetPassword)

			r.Delete("/user", userHandler.DeleteV1User)

			r.Get("/profile", profileHandler.GetV1Profile)
			r.Put("/profile", profileHandler.PutV1Profile)
			r.Post("/profile/reset", profileHandler.PostV1Reset)

			r.Get("/ingredients/autocomplete", ingredientHandler.GetAutocomplete)

			r.Get("/products/{barcode}", productHandler.GetByBarcode)
			r.Post("/products/ocr/draft", productHandler.PostOCRDraft)

			r.Post("/manual/analyze", manualHandler.PostAnalyze)
			r.Post("/manual/custom", manualHandler.PostCustom)
			r.Post("/recipe/analyze", recipeHandler.PostAnalyze)

			r.Post("/diary/entries", diaryHandler.PostEntry)
			r.Get("/diary/today", diaryHandler.GetToday)
			r.Post("/manual/add-to-diary", diaryHandler.PostManualAddToDiary)
			r.Post("/recipe/add-to-diary", diaryHandler.PostRecipeAddToDiary)

			r.Get("/onboarding/status", onboardingHandler.GetV1Status)
		})
	})

	return r
}

func corsMiddleware() func(http.Handler) http.Handler {
	allowOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	allowAll := allowOrigins == "*"

	allowed := func(origin string) bool {
		if allowAll {
			return true
		}
		if origin == "" {
			return false
		}
		originLower := strings.ToLower(origin)
		if strings.HasPrefix(originLower, "http://localhost") || strings.HasPrefix(originLower, "http://127.0.0.1") {
			return true
		}
		if allowOrigins == "" {
			return false
		}
		for _, item := range strings.Split(allowOrigins, ",") {
			item = strings.ToLower(strings.TrimSpace(item))
			if item != "" && item == originLower {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed(origin) {
				w.Header().Set("Vary", "Origin")
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-Timezone-Offset-Minutes")
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
