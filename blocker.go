package cacheproxy

import (
	"sync"
)

var blockerFile sync.RWMutex

type Blocker struct {
	mu    sync.RWMutex
	Ports map[int]sync.RWMutex
	Files map[string]sync.RWMutex
}

var blocker *Blocker

func init() {

	blockerFile = sync.RWMutex{}

	blocker = &Blocker{
		Ports: map[int]sync.RWMutex{},
		Files: map[string]sync.RWMutex{},
	}
}

func (b *Blocker) FileLocker(file string) *sync.RWMutex {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, find := b.Files[file]; !find {
		b.Files[file] = sync.RWMutex{}
	}

	val := b.Files[file]
	val.Lock()
	return &val
}

func (b *Blocker) PortLocker(port int) *sync.RWMutex {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, find := b.Ports[port]; !find {
		b.Ports[port] = sync.RWMutex{}
	}

	val := b.Ports[port]
	val.Lock()
	return &val
}
