package networkserver

import (
	"errors"
	"sync"
)

type Manager struct {
	mu sync.RWMutex
	ns map[string]*NetworkServer
}

func NewManager() *Manager {
	return &Manager{
		ns: make(map[string]*NetworkServer),
	}
}

func (m *Manager) Get(name string) (*NetworkServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ns, exists := m.ns[name]
	return ns, exists
}

func (m *Manager) Add(name string) (*NetworkServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.ns[name]; exists {
		return nil, errors.New("Network Server already exists")
	}

	m.ns[name] = New(name)
	return m.ns[name], nil
}

func (m *Manager) List() []NetworkServerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]NetworkServerInfo, 0, len(m.ns))
	for _, ns := range m.ns {
		res = append(res, ns.GetInfo())
	}

	return res
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.ns, name)
}
