package auth

import "github.com/gin-gonic/gin"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Login(c *gin.Context) {

}

func (h *Handler) Register(c *gin.Context) {

}

func (h *Handler) Logout(c *gin.Context) {

}

func (h *Handler) Me(c *gin.Context) {

}
