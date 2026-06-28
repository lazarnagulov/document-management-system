package document

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	protected := rg.Group("/documents")
	protected.Use(h.auth)
	{
		protected.GET("/", h.List)
		protected.POST("/", h.Create)
		protected.GET("/:id", h.Get)
		protected.PATCH("/:id", h.Update)
		protected.DELETE("/:id", h.Delete)
		protected.GET("/:id/download", h.Download)
		protected.POST("/:id/query", h.Query)
	}
}
