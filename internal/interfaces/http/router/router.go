package router

import (
	"log/slog"
	"time"

	"sipon-api/internal/app/port"
	"sipon-api/internal/app/service/principal"
	"sipon-api/internal/config"
	webhandler "sipon-api/internal/interfaces/http/handler/web"
	"sipon-api/internal/interfaces/http/httperror"
	"sipon-api/internal/interfaces/http/middleware"
	"sipon-api/internal/interfaces/http/respond"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(
	webAuthHandler *webhandler.AuthHandler,
	webRolePermissionHandler *webhandler.RolePermissionHandler,
	tokenGen port.TokenGenerator,
	sessionRevocationStore port.SessionRevocationStore,
	principalBuilder *principal.Builder,
	principalCache port.PrincipalCachePort,
	rateLimiter port.RateLimiter,
	rlCfg config.RateLimitConfig,
	appEnv string,
	logger *slog.Logger,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestLogger(logger))
	r.Use(httperror.Middleware(logger))

	if rlCfg.Enabled && rateLimiter != nil {
		ipWindow := time.Duration(rlCfg.IPWindowSeconds) * time.Second
		r.Use(middleware.RateLimitByIP(rateLimiter, rlCfg.IPLimit, ipWindow))
	}

	if appEnv == "development" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		respond.OK(c, "health check", gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	web := v1.Group("/web")

	// Auth routes (public)
	webAuth := web.Group("/auth")
	if rlCfg.Enabled && rateLimiter != nil {
		authWindow := time.Duration(rlCfg.AuthWindowSeconds) * time.Second
		webAuth.Use(middleware.RateLimitByIP(rateLimiter, rlCfg.AuthLimit, authWindow))
	}
	{
		webAuth.POST("/register", webAuthHandler.Register)
		webAuth.POST("/login", webAuthHandler.Login)
		webAuth.POST("/request-otp", webAuthHandler.RequestIdentityOTP)
		webAuth.POST("/verify-otp", webAuthHandler.VerifyIdentityOTP)
		webAuth.POST("/refresh-token", webAuthHandler.RefreshToken)
		webAuth.POST("/password/forgot", webAuthHandler.ForgotPassword)
		webAuth.POST("/password/reset", webAuthHandler.ResetPassword)
	}

	// Universal session bootstrap
	sessionProtected := v1.Group("")
	sessionProtected.Use(
		middleware.JWTAuth(tokenGen, sessionRevocationStore),
		middleware.PrincipalLoader(principalBuilder, principalCache),
	)
	{
		sessionProtected.GET("/auth/session", webAuthHandler.GetSession)
		sessionProtected.POST("/auth/logout", webAuthHandler.Logout)
	}

	// Protected routes
	protectedWeb := web.Group("")
	protectedWeb.Use(
		middleware.JWTAuth(tokenGen, sessionRevocationStore),
		middleware.PrincipalLoader(principalBuilder, principalCache),
	)
	if rlCfg.Enabled && rateLimiter != nil {
		userWindow := time.Duration(rlCfg.UserWindowSeconds) * time.Second
		protectedWeb.Use(middleware.RateLimitByUser(rateLimiter, rlCfg.UserLimit, userWindow))
	}
	{
		protectedWeb.GET("/auth/me", webAuthHandler.Me)
		protectedWeb.POST("/auth/change-password", webAuthHandler.ChangePasswordLocal)
		protectedWeb.POST("/auth/set-password", webAuthHandler.SetPasswordLocal)
		protectedWeb.POST("/auth/change-email/request", webAuthHandler.RequestChangeEmail)
		protectedWeb.POST("/auth/change-email/confirm", webAuthHandler.ConfirmChangeEmail)
		protectedWeb.POST("/auth/change-phone/request", webAuthHandler.RequestChangePhone)
		protectedWeb.POST("/auth/change-phone/confirm", webAuthHandler.ConfirmChangePhone)

		rolePermission := protectedWeb.Group("/role-permission", middleware.RequireRole("superadmin", "usergod"))
		{
			rolePermission.GET("/roles", webRolePermissionHandler.ListRoles)
			rolePermission.GET("/roles/:role_id", webRolePermissionHandler.GetRole)
			rolePermission.POST("/roles", webRolePermissionHandler.CreateRole)
			rolePermission.PUT("/roles/:role_id", webRolePermissionHandler.UpdateRole)

			rolePermission.GET("/permission-keys", webRolePermissionHandler.ListPermissionKeys)
			rolePermission.POST("/roles/:role_id/permissions", webRolePermissionHandler.AssignRolePermission)
			rolePermission.DELETE("/roles/:role_id/permissions/:permission_key", webRolePermissionHandler.RevokeRolePermission)

			rolePermission.GET("/user-roles", webRolePermissionHandler.ListUserRoles)
			rolePermission.GET("/user-roles/:user_role_id", webRolePermissionHandler.GetUserRole)
			rolePermission.POST("/user-roles", webRolePermissionHandler.AssignUserRole)
			rolePermission.PUT("/user-roles/:user_role_id", webRolePermissionHandler.UpdateUserRole)
			rolePermission.POST("/user-roles/:user_role_id/deactivate", webRolePermissionHandler.DeactivateUserRole)
			rolePermission.POST("/user-roles/:user_role_id/reactivate", webRolePermissionHandler.ReactivateUserRole)
			rolePermission.DELETE("/user-roles/:user_role_id", webRolePermissionHandler.DeleteUserRole)
		}
	}

	return r
}
