package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota // select random
	RoundRobinSelectSelect
)

type Discovery interface {
	Refresh() error //refresh from remote registry
	Update(server []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r       *rand.Rand // 产生随机数
	mu      sync.RWMutex
	servers []string
	index   int //记录 Round Robin 算法已经轮询到的位置，为了避免每次从 0 开始，初始化时随机设定一个值。
}

func NewMultiServersDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServersDiscovery)(nil)

func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {

	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelectSelect:
		s := d.servers[d.index%n] // 服务器可以被更新，因此n必须要保证安全
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// GetAll 返回所有的服务器列表
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)

	return servers, nil
}
