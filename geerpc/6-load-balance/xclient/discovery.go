package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MultiServerDiscovery struct {
	rwMutex sync.RWMutex //protect following，即是保护servers,index
	servers []string
	index   int
}

// 使用例子： NewMultiServerDiscovery([]string{"tcp@127.0.0.1:100","tcp@127.0.0.1:2100"})
func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
	}

	d.index = rand.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

func (d *MultiServerDiscovery) Update(servers []string) error {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.servers = servers
	return nil
}

func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	//这里不能用d.rwMutex.RLock(),因为d.index有更新
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}

	switch mode {
	case RandomSelect:
		return d.servers[rand.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
	// return a copy of d.servers
	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
