package cacheproxy

import (
	"context"
)

func Server(ctx context.Context, cfg *Config) {
	if cfg.FileName == "" {
		cfg.DynamyFileName = true
	}
	run(ctx, cfg)
}
