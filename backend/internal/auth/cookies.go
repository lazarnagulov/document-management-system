package auth

import (
	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie(AccessTokenCookie, accessToken,
		int(token.AccessTokenDuration.Seconds()),
		"/", "", true, true) // httpOnly=true

	c.SetCookie(refreshTokenCookie, refreshToken,
		int(token.RefreshTokenDuration.Seconds()),
		"/auth/refresh", "", true, true)
}

func clearAuthCookies(c *gin.Context) {
	c.SetCookie(AccessTokenCookie, "", -1, "/", "", true, true)
	c.SetCookie(refreshTokenCookie, "", -1, "/auth/refresh", "", true, true)
}
