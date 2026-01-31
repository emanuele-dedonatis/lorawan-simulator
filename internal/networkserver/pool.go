package networkserver

import (
	"errors"
	"sync"
)

type Pool struct {
	mu sync.RWMutex
	ns map[string]*NetworkServer
}

func NewPool() *Pool {
	return &Pool{
		ns: make(map[string]*NetworkServer),
	}
}

func (p *Pool) Get(name string) (*NetworkServer, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ns, exists := p.ns[name]
	return ns, exists
}

func (p *Pool) Add(name string) (*NetworkServer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.ns[name]; exists {
		return nil, errors.New("Network Server already exists")
	}

	p.ns[name] = New(name)
	return p.ns[name], nil
}

func (p *Pool) List() []NetworkServerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	res := make([]NetworkServerInfo, 0, len(p.ns))
	for _, ns := range p.ns {
		res = append(res, ns.GetInfo())
	}

	return res
}

func (p *Pool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.ns, name)
}
