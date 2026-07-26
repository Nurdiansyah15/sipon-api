package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	_ "sipon-api/docs"
	"sipon-api/internal/app/port"
	"sipon-api/internal/app/service/principal"
	authUsecase "sipon-api/internal/app/usecase/auth"
	rolePermissionUsecase "sipon-api/internal/app/usecase/rolepermission"
	"sipon-api/internal/config"
	"sipon-api/internal/infrastructure/cache"
	"sipon-api/internal/infrastructure/external/bcrypt"
	"sipon-api/internal/infrastructure/external/fonnte"
	extjwt "sipon-api/internal/infrastructure/external/jwt"
	"sipon-api/internal/infrastructure/external/otpgen"
	extsmtp "sipon-api/internal/infrastructure/external/smtp"
	"sipon-api/internal/infrastructure/persistence"
	webhandler "sipon-api/internal/interfaces/http/handler/web"
	"sipon-api/internal/interfaces/http/router"
	applogger "sipon-api/internal/logger"
)

// @title Sipon API
// @version 1.0
// @description API development docs untuk Sipon.
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Gunakan format Bearer <token>.

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	logger := applogger.New(cfg.App.Env, cfg.App.LogFormat)

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		logger.Error("gagal koneksi database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("database tidak bisa dijangkau", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("terhubung ke PostgreSQL")

	// ── Infrastructure: repositories ─────────────────────────────────────────
	userRepo := persistence.NewPostgresUserRepository(db)
	verifRepo := persistence.NewPostgresVerificationRepository(db)
	roleRepo := persistence.NewPostgresRoleRepository(db)
	userRoleRepo := persistence.NewPostgresUserRoleRepository(db)
	rolePermissionRepo := persistence.NewPostgresRolePermissionRepository(db)
	rolePermissionReadModel := persistence.NewPostgresRoleQuery(db)

	// ── Infrastructure: external services ────────────────────────────────────
	hasher := bcrypt.NewBcryptPasswordHasher()
	tokenGen := extjwt.NewJWTTokenGenerator(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)
	var emailSender port.EmailSender
	if cfg.SMTP.Host != "" && cfg.SMTP.Username != "" {
		emailSender = extsmtp.NewSMTPEmailSender(
			cfg.SMTP.Host, cfg.SMTP.Port,
			cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.From,
		)
		logger.Info("SMTP email sender aktif", slog.String("host", cfg.SMTP.Host))
	} else {
		logger.Warn("SMTP tidak dikonfigurasi, email sender dinonaktifkan (noop)")
		emailSender = extsmtp.NewNoopEmailSender()
	}
	smsSender := fonnte.NewSender(cfg.Fonnte.Token, cfg.Fonnte.URL)
	otpGen := otpgen.NewCryptoOTPGenerator()

	// ── Infrastructure: transactor ───────────────────────────────────────────
	transactor := persistence.NewPostgresTransactor(db)

	// ── Redis ─────────────────────────────────────────────────────────────────
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis tidak dapat dijangkau, principal cache dinonaktifkan", slog.Any("error", err))
	} else {
		logger.Info("terhubung ke Redis", slog.String("addr", cfg.Redis.Addr))
	}
	defer redisClient.Close()

	// SessionRevocationStore — enforce logout & logout-all secara real meski JWT
	// stateless (lihat port.SessionRevocationStore untuk detail pola key-nya).
	var sessionRevocationStore port.SessionRevocationStore = cache.NewRedisSessionRevocationStore(redisClient)

	// ── Application: use cases ────────────────────────────────────────────────
	loginUC := authUsecase.NewLoginUseCase(userRepo, hasher, tokenGen)
	refreshTokenUC := authUsecase.NewRefreshTokenUseCase(userRepo, tokenGen, sessionRevocationStore)
	changePasswordLocalUC := authUsecase.NewChangePasswordLocalUseCase(userRepo, hasher)
	setPasswordLocalUC := authUsecase.NewSetPasswordLocalUseCase(userRepo, hasher)
	requestIdentityOTPUC := authUsecase.NewRequestIdentityOTPUseCase(userRepo, verifRepo, otpGen, emailSender, smsSender)
	verifyIdentityOTPUC := authUsecase.NewVerifyIdentityOTPUseCase(userRepo, verifRepo)
	meUC := authUsecase.NewMeUseCase(userRepo)
	forgotPasswordUC := authUsecase.NewForgotPasswordUseCase(userRepo, verifRepo, otpGen, emailSender)
	resetPasswordUC := authUsecase.NewResetPasswordUseCase(userRepo, verifRepo, hasher)
	requestChangeIdentityUC := authUsecase.NewRequestChangeIdentityUseCase(userRepo, verifRepo, otpGen, emailSender, smsSender)
	confirmChangeIdentityUC := authUsecase.NewConfirmChangeIdentityUseCase(userRepo, verifRepo, transactor)
	getSessionUC := authUsecase.NewGetSessionUseCase(userRepo)
	logoutUC := authUsecase.NewLogoutUseCase(sessionRevocationStore, cfg.JWT.AccessTokenTTL)
	registerUC := authUsecase.NewRegisterUseCase(userRepo, verifRepo, hasher, otpGen, emailSender, smsSender, tokenGen, transactor, roleRepo, userRoleRepo)

	rolePermissionUseCases := rolePermissionUsecase.NewUseCases(rolePermissionUsecase.Dependencies{
		RoleRepo:           roleRepo,
		UserRoleRepo:       userRoleRepo,
		RolePermissionRepo: rolePermissionRepo,
		UserRepo:           userRepo,
		ReadModel:          rolePermissionReadModel,
	})

	// ── Principal builder & cache ─────────────────────────────────────────────
	principalCache := cache.NewRedisPrincipalCache(redisClient)
	rateLimiter := cache.NewRedisRateLimiter(redisClient)
	principalBuilder := principal.NewBuilder(userRepo, userRoleRepo, roleRepo, rolePermissionRepo)

	// ── Interface: HTTP handler & router ──────────────────────────────────────
	webAuthHandler := webhandler.NewAuthHandler(registerUC, loginUC, refreshTokenUC, changePasswordLocalUC, setPasswordLocalUC, requestIdentityOTPUC, verifyIdentityOTPUC, meUC, forgotPasswordUC, resetPasswordUC, requestChangeIdentityUC, confirmChangeIdentityUC, getSessionUC, logoutUC)
	webRolePermissionHandler := webhandler.NewRolePermissionHandler(rolePermissionUseCases)

	engine := router.Setup(
		webAuthHandler, webRolePermissionHandler,
		tokenGen, sessionRevocationStore, principalBuilder, principalCache,
		rateLimiter, cfg.RateLimit, cfg.App.Env, logger,
	)

	// ── Server dengan graceful shutdown ───────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server started",
			slog.String("addr", "http://0.0.0.0:"+cfg.App.Port),
			slog.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("server stopped")
}
