package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParamUUID(c *gin.Context, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		WriteError(c, http.StatusBadRequest, "invalid "+key)
		return uuid.UUID{}, false
	}
	return id, true
}
