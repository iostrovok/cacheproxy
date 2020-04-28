package cacheproxy

import (
	"context"
	"sync"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/handler"
)

type Manager struct {
	mu sync.RWMutex

	protoCfg config.Config
	LastPort int
	PortFrom int
	PortTo   int
	AllPorts []int
}

func NewManager(portFrom, portTo int, cfg *config.Config) *Manager {

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

func (m *Manager) SetCfg(cfg *config.Config) {
	m.protoCfg = *cfg
}

func (m *Manager) RunSrv(ctx context.Context, fileName string) (int, error) {

	m.mu.Lock()
	nextPort := m.AllPorts[m.LastPort]
	m.LastPort = (m.LastPort + 1) % len(m.AllPorts)
	m.mu.Unlock()

	copyCfg := &config.Config{
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

	err := handler.Start(ctx, copyCfg)
	if err != nil {
		return 0, err
	}

	return nextPort, nil
}
