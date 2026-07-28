package router

import (
	"log/slog"
	"time"

	"sipon-api/internal/app/port"
	roleconstant "sipon-api/internal/domain/role/constant"
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
	webUserManagementHandler *webhandler.UserManagementHandler,
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
		protectedWeb.GET("/auth/profile", webAuthHandler.Profile)
		protectedWeb.POST("/auth/change-password", webAuthHandler.ChangePasswordLocal)
		protectedWeb.POST("/auth/set-password", webAuthHandler.SetPasswordLocal)
		protectedWeb.POST("/auth/change-email/request", webAuthHandler.RequestChangeEmail)
		protectedWeb.POST("/auth/change-email/confirm", webAuthHandler.ConfirmChangeEmail)
		protectedWeb.POST("/auth/change-phone/request", webAuthHandler.RequestChangePhone)
		protectedWeb.POST("/auth/change-phone/confirm", webAuthHandler.ConfirmChangePhone)
		protectedWeb.PUT("/auth/profile", webAuthHandler.UpdateProfile)
		protectedWeb.GET("/auth/check-username", webAuthHandler.CheckUsername)
		protectedWeb.POST("/auth/change-username", webAuthHandler.ChangeUsername)
		protectedWeb.POST("/auth/profile/avatar/presign", webAuthHandler.AvatarPresign)
		protectedWeb.POST("/auth/profile/avatar/confirm", webAuthHandler.AvatarConfirm)
		protectedWeb.DELETE("/auth/profile/avatar", webAuthHandler.AvatarDelete)

		// ── User Management ───────────────────────────────────────────────────
		// Dibawah /api/v1/web/users. Tiap route dipasangkan guard RequirePermission
		// per spesifik plan §7 — admin (manage_users/reset_user_password/
		// deactivate_user) terlihat, superadmin/usergod punya semua.
		users := protectedWeb.Group("/users")
		{
			users.GET("", middleware.RequirePermission(string(roleconstant.PermissionManageUsers)), webUserManagementHandler.ListUsers)
			users.GET("/:user_id", middleware.RequirePermission(string(roleconstant.PermissionManageUsers)), webUserManagementHandler.GetUser)
			users.POST("", middleware.RequirePermission(string(roleconstant.PermissionManageUsers)), webUserManagementHandler.CreateUser)
			users.POST("/:user_id/reset-password", middleware.RequirePermission(string(roleconstant.PermissionResetUserPassword)), webUserManagementHandler.ResetUserPassword)
			users.POST("/:user_id/deactivate", middleware.RequirePermission(string(roleconstant.PermissionDeactivateUser)), webUserManagementHandler.DeactivateUser)
			users.POST("/:user_id/reactivate", middleware.RequirePermission(string(roleconstant.PermissionDeactivateUser)), webUserManagementHandler.ReactivateUser)
		}

		// ── Role & Permission Management ──────────────────────────────────────
		// Sebelumnya ditutup dengan blanket RequireRole("superadmin","usergod"),
		// yang membuat admin — padahal RolePermissions-nya termasuk assign_role —
		// tidak bisa mencapai route assign-role (latent bug, lihat
		// docs/plans/system-management-module.md §Context). Sekarang tiap route
		// pakai RequirePermission granular.
		rolePermission := protectedWeb.Group("/role-permission")
		{
			readRoleGuard := middleware.RequirePermission(
				string(roleconstant.PermissionManageRoles),
				string(roleconstant.PermissionManageRolePermissions),
				string(roleconstant.PermissionAssignRole),
			)
			rolePermission.GET("/roles", readRoleGuard, webRolePermissionHandler.ListRoles)
			rolePermission.GET("/roles/:role_id", readRoleGuard, webRolePermissionHandler.GetRole)
			rolePermission.GET("/permission-keys", readRoleGuard, webRolePermissionHandler.ListPermissionKeys)

			rolePermission.POST("/roles", middleware.RequirePermission(string(roleconstant.PermissionManageRoles)), webRolePermissionHandler.CreateRole)
			rolePermission.PUT("/roles/:role_id", middleware.RequirePermission(string(roleconstant.PermissionManageRoles)), webRolePermissionHandler.UpdateRole)

			rolePermission.POST("/roles/:role_id/permissions", middleware.RequirePermission(string(roleconstant.PermissionManageRolePermissions)), webRolePermissionHandler.AssignRolePermission)
			rolePermission.DELETE("/roles/:role_id/permissions/:permission_key", middleware.RequirePermission(string(roleconstant.PermissionManageRolePermissions)), webRolePermissionHandler.RevokeRolePermission)

			userRoleReadGuard := middleware.RequirePermission(
				string(roleconstant.PermissionAssignRole),
				string(roleconstant.PermissionManageUsers),
			)
			userRoleWriteGuard := middleware.RequirePermission(string(roleconstant.PermissionAssignRole))

			rolePermission.GET("/user-roles", userRoleReadGuard, webRolePermissionHandler.ListUserRoles)
			rolePermission.GET("/user-roles/:user_role_id", userRoleReadGuard, webRolePermissionHandler.GetUserRole)
			rolePermission.POST("/user-roles", userRoleWriteGuard, webRolePermissionHandler.AssignUserRole)
			rolePermission.PUT("/user-roles/:user_role_id", userRoleWriteGuard, webRolePermissionHandler.UpdateUserRole)
			rolePermission.POST("/user-roles/:user_role_id/deactivate", userRoleWriteGuard, webRolePermissionHandler.DeactivateUserRole)
			rolePermission.POST("/user-roles/:user_role_id/reactivate", userRoleWriteGuard, webRolePermissionHandler.ReactivateUserRole)
			rolePermission.DELETE("/user-roles/:user_role_id", userRoleWriteGuard, webRolePermissionHandler.DeleteUserRole)

			rolePermission.GET("/roles/:role_id/scopes", readRoleGuard, webRolePermissionHandler.ListRoleScopes)
			rolePermission.POST("/roles/:role_id/scopes", middleware.RequirePermission(string(roleconstant.PermissionManageRolePermissions)), webRolePermissionHandler.AssignRoleScope)
			rolePermission.DELETE("/roles/:role_id/scopes/:scope_id", middleware.RequirePermission(string(roleconstant.PermissionManageRolePermissions)), webRolePermissionHandler.RemoveRoleScope)
		}
	}

	return r
}
