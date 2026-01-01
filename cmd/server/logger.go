package main

import (
	"go.uber.org/zap"
)

func newLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
