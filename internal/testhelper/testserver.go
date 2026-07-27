// Package testhelper menyediakan utilitas untuk HTTP integration test.
// MustStartTestServer membangun server auth + role-permission menggunakan
// testcontainer PostgreSQL dan implementasi no-op/in-memory untuk layanan eksternal.
package testhelper

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"sipon-api/internal/app/port"
	"sipon-api/internal/app/service/principal"
	authUsecase "sipon-api/internal/app/usecase/auth"
	rolePermissionUsecase "sipon-api/internal/app/usecase/rolepermission"
	userManagementUsecase "sipon-api/internal/app/usecase/usermanagement"
	"sipon-api/internal/config"
	"sipon-api/internal/infrastructure/external/bcrypt"
	extjwt "sipon-api/internal/infrastructure/external/jwt"
	"sipon-api/internal/infrastructure/external/otpgen"
	extsmtp "sipon-api/internal/infrastructure/external/smtp"
	"sipon-api/internal/infrastructure/persistence"
	webhandler "sipon-api/internal/interfaces/http/handler/web"
	"sipon-api/internal/interfaces/http/router"
)

// TestServer membungkus Gin engine dan *sql.DB untuk HTTP integration test.
type TestServer struct {
	DB       *sql.DB
	Engine   *gin.Engine
	TokenGen port.TokenGenerator
	cleanup  func()
}

// MustStartTestServer membangun TestServer lengkap untuk fitur auth + authorization
// (role/user-role). Memanggil testcontainer PostgreSQL, menjalankan migrasi,
// menyemai role wajib, dan menginisialisasi semua dependency dengan
// no-op/in-memory untuk layanan eksternal.
//
// Kembalikan TestServer dan fungsi cleanup. Panggil cleanup() di akhir TestMain.
func MustStartTestServer() (*TestServer, func()) {
	gin.SetMode(gin.TestMode)

	db, dbCleanup := MustStartTestDB()

	// Seed role wajib — permission sudah dipetakan lewat constant
	// (lihat internal/domain/role/constant/permission_constant.go), jadi tidak
	// perlu seed permission/role_permission terpisah.
	ctx := context.Background()
	MustSeedNamedRole(ctx, nil, db, "member", "system", "global")
	MustSeedNamedRole(ctx, nil, db, "superadmin", "system", "global")
	MustSeedNamedRole(ctx, nil, db, "usergod", "system", "global")
	MustSeedNamedRole(ctx, nil, db, "admin", "system", "global")

	// ── Infrastructure: repositories ────────────────────────────────────────
	userRepo := persistence.NewPostgresUserRepository(db)
	verifRepo := persistence.NewPostgresVerificationRepository(db)
	roleRepo := persistence.NewPostgresRoleRepository(db)
	userRoleRepo := persistence.NewPostgresUserRoleRepository(db)
	rolePermissionRepo := persistence.NewPostgresRolePermissionRepository(db)
	rolePermissionReadModel := persistence.NewPostgresRoleQuery(db)
	userReadModel := persistence.NewPostgresUserQuery(db)
	transactor := persistence.NewPostgresTransactor(db)

	// ── Infrastructure: external services (no-op/in-memory untuk test) ─────
	hasher := bcrypt.NewBcryptPasswordHasher()
	tokenGen := extjwt.NewJWTTokenGenerator("test-jwt-secret-key-for-testing-only", 24*time.Hour, 30*24*time.Hour)
	sessionRevocationStore := newFakeSessionRevocationStore()
	emailSender := extsmtp.NewNoopEmailSender()
	smsSender := noopSMSSender{}
	otpGenerator := otpgen.NewCryptoOTPGenerator()

	// ── Application: use cases ──────────────────────────────────────────────
	loginUC := authUsecase.NewLoginUseCase(userRepo, hasher, tokenGen)
	refreshTokenUC := authUsecase.NewRefreshTokenUseCase(userRepo, tokenGen, sessionRevocationStore)
	changePasswordLocalUC := authUsecase.NewChangePasswordLocalUseCase(userRepo, hasher)
	setPasswordLocalUC := authUsecase.NewSetPasswordLocalUseCase(userRepo, hasher)
	requestIdentityOTPUC := authUsecase.NewRequestIdentityOTPUseCase(userRepo, verifRepo, otpGenerator, emailSender, smsSender)
	verifyIdentityOTPUC := authUsecase.NewVerifyIdentityOTPUseCase(userRepo, verifRepo)
	meUC := authUsecase.NewMeUseCase(userRepo)
	forgotPasswordUC := authUsecase.NewForgotPasswordUseCase(userRepo, verifRepo, otpGenerator, emailSender)
	resetPasswordUC := authUsecase.NewResetPasswordUseCase(userRepo, verifRepo, hasher)
	requestChangeIdentityUC := authUsecase.NewRequestChangeIdentityUseCase(userRepo, verifRepo, otpGenerator, emailSender, smsSender)
	confirmChangeIdentityUC := authUsecase.NewConfirmChangeIdentityUseCase(userRepo, verifRepo, transactor)
	getSessionUC := authUsecase.NewGetSessionUseCase(userRepo)
	getProfileUC := authUsecase.NewGetProfileUseCase(userRepo)
	logoutUC := authUsecase.NewLogoutUseCase(sessionRevocationStore, 24*time.Hour)
	registerUC := authUsecase.NewRegisterUseCase(userRepo, verifRepo, hasher, otpGenerator, emailSender, smsSender, tokenGen, transactor, roleRepo, userRoleRepo)

	rolePermissionUseCases := rolePermissionUsecase.NewUseCases(rolePermissionUsecase.Dependencies{
		RoleRepo:           roleRepo,
		UserRoleRepo:       userRoleRepo,
		RolePermissionRepo: rolePermissionRepo,
		UserRepo:           userRepo,
		ReadModel:          rolePermissionReadModel,
	})

	userManagementUseCases := userManagementUsecase.NewUseCases(userManagementUsecase.Dependencies{
		UserRepo:  userRepo,
		ReadModel: userReadModel,
		Hasher:    hasher,
	})

	// ── Principal builder & cache ───────────────────────────────────────────
	principalCache := noopPrincipalCache{}
	principalBuilder := principal.NewBuilder(userRepo, userRoleRepo, roleRepo, rolePermissionRepo)

	// ── HTTP handlers & router ──────────────────────────────────────────────
	webAuthHandler := webhandler.NewAuthHandler(registerUC, loginUC, refreshTokenUC, changePasswordLocalUC, setPasswordLocalUC, requestIdentityOTPUC, verifyIdentityOTPUC, meUC, forgotPasswordUC, resetPasswordUC, requestChangeIdentityUC, confirmChangeIdentityUC, getSessionUC, logoutUC, getProfileUC)
	webRolePermHandler := webhandler.NewRolePermissionHandler(rolePermissionUseCases)
	webUserManagementHandler := webhandler.NewUserManagementHandler(userManagementUseCases)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	rlCfg := config.RateLimitConfig{Enabled: false}

	engine := router.Setup(
		webAuthHandler, webRolePermHandler, webUserManagementHandler,
		tokenGen, sessionRevocationStore, principalBuilder, principalCache,
		nil, rlCfg, "test", logger,
	)

	srv := &TestServer{
		DB:       db,
		Engine:   engine,
		TokenGen: tokenGen,
		cleanup:  dbCleanup,
	}
	return srv, dbCleanup
}

// ── HTTP helper methods ────────────────────────────────────────────────────────

func (s *TestServer) do(method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			panic("testhelper: marshal request body: " + err.Error())
		}
		bodyBytes = b
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	if err != nil {
		panic("testhelper: new request: " + err.Error())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	s.Engine.ServeHTTP(w, req)
	return w
}

func (s *TestServer) POST(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodPost, path, body, mergeHeaders(headers))
}

func (s *TestServer) GET(path string, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodGet, path, nil, mergeHeaders(headers))
}

func (s *TestServer) PATCH(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodPatch, path, body, mergeHeaders(headers))
}

func (s *TestServer) PUT(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodPut, path, body, mergeHeaders(headers))
}

func (s *TestServer) DELETE(path string, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodDelete, path, nil, mergeHeaders(headers))
}

func (s *TestServer) DELETEBody(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	return s.do(http.MethodDelete, path, body, mergeHeaders(headers))
}

// BearerHeader mengembalikan map header Authorization: Bearer <token>.
func BearerHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// DecodeBody melakukan unmarshal data JSON dari field "data" pada response body.
func DecodeBody(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&wrapper), "decode response body")
	if out != nil {
		require.NoError(t, json.Unmarshal(wrapper.Data, out), "decode data field")
	}
}

func mergeHeaders(headers []map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, h := range headers {
		for k, v := range h {
			merged[k] = v
		}
	}
	return merged
}
