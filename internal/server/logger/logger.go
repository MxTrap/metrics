package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

func (l *Logger) LoggerInterceptor(
	ctx context.Context,
	req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	l.Logger.Infoln(
		"full method", info.FullMethod,
		"duration", duration,
		"status", resp,
		"err", err,
	)

	return resp, err
}
