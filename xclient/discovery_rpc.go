package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type RPCRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewRPCRegistryDiscovery(registerAddr string, timeout time.Duration) *RPCRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}

	d := &RPCRegistryDiscovery{
		MultiServersDiscovery: NewMultiServersDiscovery(make([]string, 0)),
		registry:              registerAddr,
		timeout:               timeout,
	}
	return d
}

func (d *RPCRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 还没有过期
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}

	log.Println("rpc registry: refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)

	if err != nil {
		log.Println("rpc registry refresh err：", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-rpc-server"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *RPCRegistryDiscovery) Update(server []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = server
	d.lastUpdate = time.Now()
	return nil
}

func (d *RPCRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}

func (d *RPCRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}
