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
	mx        sync.RWMutex
	conns     map[string]*SQL
	deleteOld bool
}

var pull *Pull

func New(ctx ...context.Context) *Pull {

	out := &Pull{
		mx:    sync.RWMutex{},
		conns: map[string]*SQL{},
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

func Init(ctx ...context.Context) {
	globalPullMutex.Lock()
	defer globalPullMutex.Unlock()

	if pull == nil {
		pull = New(ctx...)
	}
}

// --------------------------------
func Close() error {
	return pull.Close()
}

func Add(fileName string) (*SQL, error) {
	return pull.Add(fileName)
}

func Get(fileName string) (*SQL, error) {
	return pull.Get(fileName)
}

func Upsert(fileName, args string, unit *store.StoreUnit) error {
	return pull.Upsert(fileName, args, unit)
}

func Select(fileName, args string) (*store.StoreUnit, error) {
	return pull.Select(fileName, args)
}

func DeleteOldFromNow() error {
	return pull.DeleteOldFromNow()
}

// -----------------------------------

func (p *Pull) DeleteOldFromNow() error {
	p.deleteOld = true

	for _, c := range p.conns {
		if err := c.fixTimeForCut(); err != nil {
			return err
		}
	}
	return nil
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

	if !p.deleteOld {
		return 0, nil
	}

	total := int64(0)

	for _, c := range p.conns {
		if c == nil {
			continue
		}

		deleted, err := c.DeleteOld()
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

	if p.deleteOld {
		err = c.fixTimeForCut()
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
func (p *Pull) Upsert(fileName, args string, unit *store.StoreUnit) error {
	c, err := p.Get(fileName)
	if err != nil {
		return err
	}
	return c.Upsert(args, unit)
}

func (p *Pull) Select(fileName, args string) (*store.StoreUnit, error) {
	c, err := p.Get(fileName)
	if err != nil {
		return nil, err
	}
	return c.Select(args)
}
