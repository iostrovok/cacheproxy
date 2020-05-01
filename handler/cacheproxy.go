package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/sqlite"
)

func Start(ctx context.Context, cfg *config.Config) error {

	err := cfg.Init()
	if err != nil {
		return err
	}

	sqlite.Init(cfg.SessionMode, ctx)

	// server wants to serve itself port
	portBlocker.Lock(cfg.Port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return err
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler(cfg, w, r)
		}),
	}

	go func(cfg *config.Config, server *http.Server, listener net.Listener) {

		ch := make(chan error, 1)
		if cfg.Scheme == "https" {
			select {
			case <-ctx.Done():
				// nothing
				logPrintf(cfg, "Done! %d", cfg.Port)
			case ch <- server.ServeTLS(listener, cfg.PemPath, cfg.KeyPath):
				// nothing
			}
		} else {
			select {
			case <-ctx.Done():
				logPrintf(cfg, "Done! %d", cfg.Port)
			case ch <- server.Serve(listener):
				// nothing
			}
		}

		logError(cfg, <-ch)
		logPrintf(cfg, "Force close server. Port: %d, error: %v", cfg.Port, server.Close())
		portBlocker.Unlock(cfg.Port)
	}(cfg, server, listener)

	return nil
}
