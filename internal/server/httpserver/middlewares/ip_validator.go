package middlewares

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
)

func IPValidator(cidr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cidr == "" {
			c.Next()
			return
		}

		xRealIP := c.Request.Header.Get("X-Real-IP")
		if xRealIP == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(xRealIP)
		if ip == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !ipNet.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
