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

func (p *Pool) Get(name string) (*NetworkServer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ns, exists := p.ns[name]

	if !exists {
		return nil, errors.New("network server not found")
	}

	return ns, nil
}

func (p *Pool) Add(name string) (*NetworkServer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.ns[name]; exists {
		return nil, errors.New("network server already exists")
	}

	p.ns[name] = New(name)
	return p.ns[name], nil
}

func (p *Pool) List() []*NetworkServer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	servers := make([]*NetworkServer, 0, len(p.ns))
	for _, ns := range p.ns {
		servers = append(servers, ns)
	}
	return servers
}

func (p *Pool) Remove(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.ns[name]; !exists {
		return errors.New("network server not found")
	}

	delete(p.ns, name)

	return nil
}
