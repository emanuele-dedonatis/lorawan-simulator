package networkserver

import (
	"errors"
	"log"
	"sync"

	"github.com/brocaar/lorawan"
	"github.com/emanuele-dedonatis/lorawan-simulator/internal/integration"
)

type Pool struct {
	mu                sync.RWMutex
	ns                map[string]*NetworkServer
	broadcastUplink   chan lorawan.PHYPayload
	broadcastDownlink chan lorawan.PHYPayload
}

func NewPool() *Pool {
	p := &Pool{
		ns:                make(map[string]*NetworkServer),
		broadcastUplink:   make(chan lorawan.PHYPayload),
		broadcastDownlink: make(chan lorawan.PHYPayload),
	}

	go p.broadcastUplinkWorker()
	go p.broadcastDownlinkWorker()

	return p
}

func (p *Pool) broadcastUplinkWorker() {
	for uplink := range p.broadcastUplink {
		p.mu.RLock()
		for _, ns := range p.ns {
			log.Printf("[pool] propagating uplink to network server %s", ns.name)
			go ns.ForwardUplink(uplink)
		}
		p.mu.RUnlock()
	}
}

func (p *Pool) broadcastDownlinkWorker() {
	for downlink := range p.broadcastDownlink {
		p.mu.RLock()
		for _, ns := range p.ns {
			log.Printf("[pool] propagating downlink to network server %s", ns.name)
			go ns.ForwardDownlink(downlink)
		}
		p.mu.RUnlock()
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

func (p *Pool) Add(name string, config integration.NetworkServerConfig) (*NetworkServer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.ns[name]; exists {
		return nil, errors.New("network server already exists")
	}

	p.ns[name] = New(name, config, p.broadcastUplink, p.broadcastDownlink)

	// Sync gateways and devices
	err := p.ns[name].Sync()
	if err != nil {
		return nil, err
	}

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
