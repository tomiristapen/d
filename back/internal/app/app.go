package app

import (
	"context"
	"fmt"
	"net/http"

	"back/internal/modules/auth"
	"back/internal/modules/diary"
	"back/internal/modules/ingredient"
	"back/internal/modules/input"
	"back/internal/modules/onboarding"
	"back/internal/modules/product"
	"back/internal/modules/profile"
	"back/internal/modules/recipe"
	"back/internal/modules/user"
	"back/internal/platform/config"
	"back/internal/platform/database"
	"back/internal/platform/httpapi"
	"back/internal/platform/jwtutil"
	"back/internal/platform/mailer"
	"back/internal/platform/migrations"
)

func Run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrations.Run(ctx, db, "internal/platform/migrations"); err != nil {
		return err
	}

	tokenManager := jwtutil.NewManager(cfg.JWTIssuer, cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	var emailSender mailer.Sender = mailer.NewLogSender(cfg.MailerFrom, cfg.MailerBaseURL)
	if cfg.SMTPHost != "" && cfg.SMTPPort != "" {
		smtpSender, err := mailer.NewSMTPSender(cfg.MailerFrom, cfg.MailerBaseURL, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
		if err != nil {
			return err
		}
		emailSender = smtpSender
	}

	authService := auth.NewService(
		auth.NewPostgresUserRepository(db),
		auth.NewPostgresVerificationRepository(db),
		auth.NewPostgresRefreshTokenRepository(db),
		auth.NewPostgresTxRunner(db),
		tokenManager,
		emailSender,
		auth.NewGoogleIDTokenVerifier(cfg.GoogleClientID),
	)
	profileService := profile.NewService(profile.NewPostgresRepository(db))
	ingredientService := ingredient.NewService(ingredient.NewPostgresRepository(db))
	productService := product.NewService(
		product.NewPostgresRepository(db),
		product.NewOpenFoodFactsClient(""),
		product.NewPostgresBaseProductRepository(db),
	)
	manualInputService := input.NewService(productService)
	recipeService := recipe.NewService(manualInputService)
	diaryService := diary.NewService(diary.NewPostgresRepository(db))
	userService := user.NewService(user.NewPostgresRepository(db))
	onboardingService := onboarding.NewService(onboarding.NewPostgresRepository(db))

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	return http.ListenAndServe(addr, httpapi.NewRouter(
		auth.NewHandler(authService),
		user.NewHandler(userService),
		profile.NewHandler(profileService),
		ingredient.NewHandler(ingredientService),
		product.NewHandler(productService),
		input.NewHandler(manualInputService),
		recipe.NewHandler(recipeService),
		diary.NewHandler(diaryService, manualInputService, recipeService),
		onboarding.NewHandler(onboardingService),
		tokenManager,
	))
}
