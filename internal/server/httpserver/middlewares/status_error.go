package middlewares

import (
	"errors"
	"fmt"
	"github.com/MxTrap/metrics/internal/server/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func StatusErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		if errors.Is(err, models.ErrNotFoundMetric) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if errors.Is(err, models.ErrUnknownMetricType) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if errors.Is(err, models.ErrWrongMetricValue) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		fmt.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)

	}
}
