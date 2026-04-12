package lib

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flora-hive/internal/domain/models"
)

// ParseBody binds JSON into body.
func ParseBody[K any](c *gin.Context, body K) (K, error) {
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, models.CreateErrorWrapped("invalid body", models.CreateErrorWithContext(err)))
		return body, err
	}
	return body, nil
}

// ParseParamInt parses an integer path param.
func ParseParamInt(c *gin.Context, param string) (int, error) {
	idStr := c.Param(param)
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, models.CreateErrorWrapped("invalid param, should be integer", models.CreateErrorWithContext(err)))
		return idInt, err
	}
	return idInt, nil
}

// ParseParamUUID parses a UUID path param.
func ParseParamUUID(c *gin.Context, param string) (uuid.UUID, error) {
	idStr := c.Param(param)
	idUUID, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, models.CreateErrorWrapped("invalid param, should be uuid", models.CreateErrorWithContext(err)))
		return idUUID, err
	}
	return idUUID, nil
}
