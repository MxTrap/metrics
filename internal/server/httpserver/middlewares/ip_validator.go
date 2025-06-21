package middlewares

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
)

func IpValidator(cidr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cidr == "" {
			c.Next()
			return
		}

		xRealIp := c.Request.Header.Get("X-Real-IP")
		if xRealIp == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		ip := net.ParseIP(xRealIp)
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
