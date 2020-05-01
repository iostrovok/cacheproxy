package sqlite

import (
	"context"
	"sync"

	"github.com/iostrovok/cacheproxy/store"
)

var globalPullMutex sync.Mutex

func init() {
	globalPullMutex = sync.Mutex{}
}

type Pull struct {
	mx          sync.RWMutex
	conns       map[string]*SQL
	sessionMode bool

	requested map[string]bool
}

// global variable
var pull *Pull

func New(sessionMode bool, ctx ...context.Context) *Pull {

	out := &Pull{
		mx:          sync.RWMutex{},
		conns:       map[string]*SQL{},
		requested:   map[string]bool{},
		sessionMode: sessionMode,
	}

	if len(ctx) > 0 {
		go func(ctx context.Context) {
			<-ctx.Done()
			out.DeleteOld()
			out.Close()
		}(ctx[0])
	}

	return out
}

func Init(sessionMode bool, ctx ...context.Context) {
	globalPullMutex.Lock()
	defer globalPullMutex.Unlock()

	if pull == nil {
		pull = New(sessionMode, ctx...)
	}
}

// --------------------------------
func Close() error {
	return pull.Close()
}

func Upsert(fileName, id string, unit *store.Item) error {
	return pull.Upsert(fileName, id, unit)
}

func Select(fileName, id string) (*store.Item, error) {
	return pull.Select(fileName, id)
}

func (p *Pull) Close() error {
	p.mx.Lock()
	defer p.mx.Unlock()

	for _, c := range p.conns {
		if c != nil {
			if err := c.Close(); err != nil {
				return err
			}
		}
		c = nil
	}

	// finally clean
	p.conns = map[string]*SQL{}
	return nil
}

func (p *Pull) DeleteOld() (int64, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if !p.sessionMode {
		return 0, nil
	}

	total := int64(0)
	for _, c := range p.conns {
		if c == nil {
			continue
		}

		deleted, err := c.DeleteOld(p.requested)
		if err != nil {
			return 0, err
		}
		total += deleted
	}

	return total, nil
}

// Add creates new connection and adds to pull
func (p *Pull) Add(fileName string) (*SQL, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if old, find := p.conns[fileName]; find {
		return old, nil
	}

	c, err := conn(fileName)
	if err == nil {
		p.conns[fileName] = c
	}

	return c, err
}

// Add creates new connection and adds to pull
func (p *Pull) Get(fileName string) (*SQL, error) {
	p.mx.RLock()
	old, find := p.conns[fileName]
	if find {
		p.mx.RUnlock()
		return old, nil
	}
	p.mx.RUnlock()

	return p.Add(fileName)
}

// Upsert just inserts or update one record
func (p *Pull) Upsert(fileName, id string, unit *store.Item) error {
	c, err := p.Get(fileName)
	if err != nil {
		return err
	}

	if p.sessionMode {
		p.mx.Lock()
		p.requested[id] = true
		p.mx.Unlock()
	}

	return c.Upsert(id, unit)
}

func (p *Pull) Select(fileName, id string) (*store.Item, error) {
	c, err := p.Get(fileName)
	if err != nil {
		return nil, err
	}
	return c.Select(id)
}
