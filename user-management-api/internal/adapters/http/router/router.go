package router

import (
	"github.com/gin-gonic/gin"

	"sevensolutions-backend/internal/adapters/http/handler"
	"sevensolutions-backend/internal/adapters/http/middleware"
	"sevensolutions-backend/pkg/jwt"
)

func New(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, jwtService *jwt.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	public := r.Group("/api/auth")
	public.POST("/register", authHandler.Register)
	public.POST("/login", authHandler.Login)

	protected := r.Group("/api/users")
	protected.Use(middleware.Auth(jwtService))
	// logic เดียวกับ /api/auth/register แค่เปิด route ให้ผ่าน auth ด้วย
	protected.POST("", authHandler.Register)
	protected.GET("", userHandler.List)
	protected.GET("/:id", userHandler.GetByID)
	protected.PUT("/:id", userHandler.Update)
	protected.DELETE("/:id", userHandler.Delete)

	return r
}
