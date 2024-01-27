package xclient

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GeeRegistryDiscovery struct {
	*MultiServerDiscovery
	registryAddr string
	timeout      time.Duration //服务列表的过期时间
	lastUpdate   time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewGeeRegistryDiscovery(registerAddr string, timeout time.Duration) *GeeRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	return &GeeRegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registryAddr:         registerAddr,
		timeout:              timeout,
	}
}

func (d *GeeRegistryDiscovery) Update(servers []string) error {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

// 刷新，有了注册中心，在客户端每次获取服务实例时候，需要刷新注册中心的保存的服务实例
func (d *GeeRegistryDiscovery) Refresh() error {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	//注册中心保存的服务实例还没超时，不用更新
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	httpClient := http.Client{Timeout: time.Second * 10} //http客户端最好有个超时
	resp, err := httpClient.Get(d.registryAddr)
	if err != nil {
		fmt.Println("rpc registry refresh err:", err)
		return err
	}

	defer resp.Body.Close()
	servers := strings.Split(resp.Header.Get("X-rpc-Servers"), ",")
	fmt.Println("servers:", servers)
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		//返回一个string类型，并将最前面和最后面的ASCII定义的空格去掉，中间的空格不会去掉
		s := strings.TrimSpace(server)
		if s != "" {
			d.servers = append(d.servers, s)
		}
	}

	d.lastUpdate = time.Now()
	return nil
}

func (d *GeeRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	//d.Get(mode) 表示调用的是(GeeRegistryDiscovery).Get
	return d.MultiServerDiscovery.Get(mode) //d.MultiServerDiscovery是调用MultiServerDiscovery的Get()
}

func (d *GeeRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServerDiscovery.GetAll()
}
