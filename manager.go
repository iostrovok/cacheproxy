package cacheproxy

import (
	"context"
	"path/filepath"
	"regexp"
	"sync"
)

type Manager struct {
	mu sync.RWMutex

	protoCfg Config
	LastPort int
	PortFrom int
	PortTo   int
	useHttps bool
}

var re *regexp.Regexp = regexp.MustCompile(`[^-_a-zA-Z0-9]+`)

func NewManager(portFrom, portTo int, cfg *Config) *Manager {
	return &Manager{
		LastPort: portFrom,
		PortFrom: portFrom,
		PortTo:   portTo,
		protoCfg: *cfg,
	}
}

func (m *Manager) SetCfg(cfg *Config) {
	m.protoCfg = *cfg
}

func (m *Manager) RunSrv(ctxIn context.Context, fileName string) (int, context.CancelFunc) {

	ctx, cancel := context.WithCancel(ctxIn)

	m.mu.Lock()
	delta := m.PortTo - m.PortFrom
	if delta < 1 {
		delta = 1
		m.LastPort = m.PortFrom
	}
	copyCfg := &Config{
		URL:       m.protoCfg.URL,
		Scheme:    m.protoCfg.Scheme,
		PemPath:   m.protoCfg.PemPath,
		KeyPath:   m.protoCfg.KeyPath,
		StorePath: m.protoCfg.StorePath,
		FileName:  fileName,
		Verbose:   m.protoCfg.Verbose,
		ForceSave: m.protoCfg.ForceSave,
		Port:      m.PortFrom + (m.LastPort-m.PortFrom)%delta,
	}
	m.LastPort++
	run(ctx, copyCfg)
	m.mu.Unlock()

	return copyCfg.Port, cancel
}

type Config struct {
	URL, Scheme      string
	Port             int
	PemPath, KeyPath string
	StorePath        string
	FileName         string
	Verbose          bool
	ForceSave        bool
	DynamyFileName   bool
}

func (cfg Config) File(urlPath string) string {

	if !cfg.DynamyFileName && cfg.FileName != "" {
		return filepath.Join(cfg.StorePath, cfg.FileName)
	}
	
	urlPath = re.ReplaceAllString(urlPath, "") + ".zip"
	return filepath.Join(cfg.StorePath, urlPath)
}
