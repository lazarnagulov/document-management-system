package httputil

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func QueryInt(c *gin.Context, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return def
	}
	return val
}
