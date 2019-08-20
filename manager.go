package cacheproxy

import (
	"context"
	"net/url"
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
	AllPorts []int
}

var re *regexp.Regexp = regexp.MustCompile(`[^-_a-zA-Z0-9]+`)

func NewManager(portFrom, portTo int, cfg *Config) *Manager {

	if portTo-portFrom < 0 {
		portTo = portFrom
	}

	allPorts := []int{}
	for i := portFrom; i <= portTo; i++ {
		allPorts = append(allPorts, i)
	}

	return &Manager{
		LastPort: 0,
		PortFrom: portFrom,
		PortTo:   portTo,
		protoCfg: *cfg,
		AllPorts: allPorts,
	}
}

func (m *Manager) SetCfg(cfg *Config) {
	m.protoCfg = *cfg
}

func (m *Manager) RunSrv(ctx context.Context, fileName string) (int, error) {

	m.mu.Lock()
	nextPort := m.AllPorts[m.LastPort]
	m.LastPort = (m.LastPort + 1) % len(m.AllPorts)
	m.mu.Unlock()

	copyCfg := &Config{
		Scheme:    m.protoCfg.Scheme,
		Host:      m.protoCfg.Host,
		PemPath:   m.protoCfg.PemPath,
		KeyPath:   m.protoCfg.KeyPath,
		StorePath: m.protoCfg.StorePath,
		FileName:  fileName,
		Verbose:   m.protoCfg.Verbose,
		ForceSave: m.protoCfg.ForceSave,
		Port:      nextPort,
	}

	err := run(ctx, copyCfg)
	if err != nil {
		return 0, err
	}

	return nextPort, nil
}

type Config struct {
	Host             string
	Scheme           string
	Port             int
	PemPath, KeyPath string
	StorePath        string
	FileName         string
	Verbose          bool
	ForceSave        bool
	DynamyFileName   bool
	URL              *url.URL
}

func (cfg *Config) init() (err error) {
	cfg.URL, err = url.Parse(cfg.Host)
	return
}

func (cfg *Config) File(urlPath string) string {

	if !cfg.DynamyFileName && cfg.FileName != "" {
		return filepath.Join(cfg.StorePath, cfg.FileName)
	}

	urlPath = re.ReplaceAllString(urlPath, "") + ".zip"
	return filepath.Join(cfg.StorePath, urlPath)
}
