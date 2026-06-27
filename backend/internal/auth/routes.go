package auth

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/register", h.Register)
		auth.POST("/refresh", h.Refresh)
	}

	protected := rg.Group("/")
	protected.Use(h.auth)
	{
		protected.GET("/auth/me", h.Me)
	}
}
