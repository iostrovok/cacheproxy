package cacheproxy

import (
	"context"
)

func Server(ctx context.Context, cfg *Config) error {
	if cfg.FileName == "" {
		cfg.DynamyFileName = true
	}
	return run(ctx, cfg)
}
