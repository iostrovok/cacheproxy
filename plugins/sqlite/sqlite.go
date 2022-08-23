package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/iostrovok/cacheproxy/cerrors"
	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/plugins"
	"github.com/iostrovok/cacheproxy/sqlite"
)

type Sqlite struct {
	storePath string
}

func New(ctx context.Context, cfg *config.Config) plugins.IPlugin {
	sqlite.Init(cfg.SessionMode, ctx)
	return &Sqlite{
		storePath: cfg.StorePath,
	}
}

func (s *Sqlite) SetVersion(_ string) error {
	return errors.Wrap(cerrors.PluginHasNoVersion, "sqlite plugin")
}

func (s *Sqlite) PreloadByVersion() error {
	return errors.Wrap(cerrors.PluginHasNoVersion, "sqlite plugin")
}

func (s *Sqlite) Read(fileName, key string) ([]byte, error) {
	store, err := sqlite.Select(s.fullFileName(fileName), key)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}

	return store, err
}

func (s *Sqlite) Save(fileName, key string, data []byte) error {
	return sqlite.Upsert(s.fullFileName(fileName), key, data)
}

func (s *Sqlite) fullFileName(file string) string {
	if file == "" {
		return strings.TrimSuffix(filepath.Join(s.storePath, " "), " ") + ".db"
	}
	return filepath.Join(s.storePath, file) + ".db"
}
