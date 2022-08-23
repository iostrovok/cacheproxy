package pg

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq"

	"github.com/iostrovok/cacheproxy/plugins"
)

/*



CREATE TABLE IF NOT EXISTS public.dbfiles
(
    id SERIAL,
    file_name character varying(40) COLLATE pg_catalog."default" NOT NULL,
    key character varying(40) COLLATE pg_catalog."default" NOT NULL,
    version character varying(40) COLLATE pg_catalog."default" NOT NULL,
    data bytea,
    CONSTRAINT pkey PRIMARY KEY (id),
    CONSTRAINT uxk UNIQUE (file_name, key, version) WITH (FILLFACTOR=100)
);

*/

type Config struct {
	// Table options
	Table      string // schema name + table.name
	FileCol    string //  character varying field for file name
	KeyCol     string //  character varying field for key
	ValCol     string //  BYTEA (binary) field for value
	VersionCol string //  character varying field for version

	//
	Version string // value for version for current request series. Keep it empty if you don't use versions.

	// cache options
	UseCache   bool // use or not use cache
	UsePreload bool // loads all data for the version from DB and saves them to cache. Useful for tests

	// Human-readable file name, it's false by default and plugin uses MD5 of file name.
	// The length of the file name depends on the length of the URL and can be very long.
	// Use HumanReadableFileName=true to prevent FileCol size error.
	HumanReadableFileName bool
}

type cacheItem struct {
	sum   [16]byte
	value []byte
}

type PG struct {
	sync.RWMutex

	cfg *Config
	ctx context.Context
	db  *sql.DB

	upsert string
	find   string
	cache  map[[16]byte]*cacheItem
}

func New(ctx context.Context, db *sql.DB, cfg *Config) (plugins.IPlugin, error) {
	out := &PG{
		ctx:   ctx,
		cfg:   cfg,
		db:    db,
		cache: map[[16]byte]*cacheItem{},
	}
	err := out.SetVersion(cfg.Version)

	return out, err
}

func (p *PG) SetVersion(version string) error {
	p.cfg.Version = version

	p.find = fmt.Sprintf(`
			SELECT %s FROM %s
			WHERE %s= $1 AND %s = $2 AND %s = '%s'
		`,
		p.cfg.ValCol, p.cfg.Table,
		p.cfg.FileCol, p.cfg.KeyCol, p.cfg.VersionCol, p.cfg.Version)

	p.upsert = fmt.Sprintf(`
			INSERT INTO  %s 
			(%s, %s, %s, %s) 
			VALUES($1, $2, $3, $4)
			ON CONFLICT (%s, %s, %s) DO 
			UPDATE SET
			%s = EXCLUDED.%s
		`,
		p.cfg.Table,
		p.cfg.FileCol, p.cfg.KeyCol, p.cfg.VersionCol, p.cfg.ValCol,
		p.cfg.FileCol, p.cfg.KeyCol, p.cfg.VersionCol,
		p.cfg.ValCol, p.cfg.ValCol)

	return nil
}

// PreloadByVersion loads all data for the version from DB and saves them to cache.
func (p *PG) PreloadByVersion() error {
	if !p.cfg.UseCache || !p.cfg.UsePreload {
		return nil
	}

	p.Lock()
	defer p.Unlock()

	sql := fmt.Sprintf(`
			SELECT %s, %s, %s FROM %s 
			WHERE %s = '%s'
		`, p.cfg.FileCol, p.cfg.KeyCol, p.cfg.ValCol, p.cfg.Table,
		p.cfg.VersionCol, p.cfg.Version)

	rows, err := p.db.QueryContext(p.ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var fileName, key string
		var data []byte

		if err := rows.Scan(&fileName, &key, &data); err != nil {
			return err
		}

		p.cache[cacheKey(fileName, key, p.cfg.Version)] = &cacheItem{
			md5.Sum(data), data,
		}
	}

	return nil
}

func (p *PG) findInCache(fileName, key string, sum [16]byte) bool {
	if !p.cfg.UseCache {
		return false
	}

	p.RLock()
	val, find := p.cache[cacheKey(fileName, key, p.cfg.Version)]
	p.RUnlock()
	if !find || val == nil {
		return false
	}

	return val.sum == sum
}

func (p *PG) readCache(fileName, key string) ([]byte, bool) {
	if !p.cfg.UseCache {
		return nil, false
	}

	p.RLock()
	defer p.RUnlock()

	cOut, find := p.cache[cacheKey(fileName, key, p.cfg.Version)]
	if !find {
		return nil, false
	}

	out := make([]byte, len(cOut.value))
	copy(out, cOut.value)
	return out, true
}

func (p *PG) shortFileName(fileName string) string {
	if p.cfg.HumanReadableFileName {
		return fileName
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(fileName)))
}

func (p *PG) Read(fileName, key string) ([]byte, error) {
	fileName = p.shortFileName(fileName)

	if out, find := p.readCache(fileName, key); find {
		return out, nil
	}

	out := make([]byte, 0)
	err := p.db.QueryRowContext(p.ctx, p.find, fileName, key).Scan(&out)
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}

	return out, err
}

func (p *PG) Save(fileName, key string, data []byte) error {
	fileName = p.shortFileName(fileName)

	if p.cfg.UseCache && p.findInCache(fileName, key, md5.Sum(data)) {
		return nil
	}

	_, err := p.db.ExecContext(p.ctx, p.upsert, fileName, key, p.cfg.Version, data)
	if err == nil && p.cfg.UseCache {
		p.Lock()
		p.cache[cacheKey(fileName, key, p.cfg.Version)] = &cacheItem{
			md5.Sum(data), data,
		}
		p.Unlock()
	}

	return err
}

func cacheKey(fileName, key, version string) [16]byte {
	return md5.Sum([]byte(fileName + "#--#" + key + "#--#" + version))
	//return string(s[:])
}
