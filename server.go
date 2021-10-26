package cacheproxy

import (
	"context"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/handler"
)

func Server(ctx context.Context, cfg *config.Config) error {
	if cfg.FileName == "" {
		cfg.DynamoFileName = true
	}

	return handler.Start(ctx, cfg)
}
