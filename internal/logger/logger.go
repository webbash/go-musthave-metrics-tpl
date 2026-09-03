// Package logger provides the application's structured logger.
package logger

import (
	"log"

	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %s", err)
	}
	defer logger.Sync()

	return logger.Sugar()
}
