package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

type Logger struct {
	Logger zap.SugaredLogger
}

func NewLogger() *Logger {
	sugar := zap.NewExample().Sugar()
	return &Logger{
		Logger: *sugar,
	}
}

func (l *Logger) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		l.Logger.Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"duration", duration,
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
		)
	}
}
