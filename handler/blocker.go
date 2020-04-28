package handler

import (
	"sync"
)

type PortBlocker struct {
	muLock   sync.RWMutex
	muUnlock sync.RWMutex
	Ports    map[int]*sync.RWMutex
}

var portBlocker *PortBlocker

func init() {
	portBlocker = &PortBlocker{
		Ports: map[int]*sync.RWMutex{},
	}
}

func (b *PortBlocker) Lock(port int) {
	b.muLock.Lock()
	defer b.muLock.Unlock()

	if _, find := b.Ports[port]; !find {
		b.Ports[port] = &sync.RWMutex{}
	}

	b.Ports[port].Lock()
}

func (b *PortBlocker) Unlock(port int) {
	b.muUnlock.Lock()
	defer b.muUnlock.Unlock()

	if _, find := b.Ports[port]; !find {
		return
	}

	b.Ports[port].Unlock()
}
