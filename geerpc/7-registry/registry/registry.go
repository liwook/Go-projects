package registry

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultPath    = "/_rpc_/registry"
	defaultTimeout = time.Minute * 5
)

type ServerItem struct {
	Addr  string
	start time.Time //用于心跳时间计算
}

// GeeRegistry is a simple register center
type GeeRegistry struct {
	timeout time.Duration
	mutex   sync.Mutex //protcect servers
	servers map[string]*ServerItem
}

var DefalultGeeRegister = New(defaultTimeout)

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

func (r *GeeRegistry) putServer(addr string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now() // if exists, update start time to keep alive
	}
}

func (r *GeeRegistry) aliveServers() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

// HTTP部分
func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-rpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-rpc-Servers")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr) //更新保存在注册中心的服务实例
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (r *GeeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
}

func HandleHTTP() {
	DefalultGeeRegister.HandleHTTP(defaultPath)
}

// only send once
func sendHeartbeat(registryURL, addr string) error {
	httpClient := &http.Client{Timeout: time.Second * 10}
	req, _ := http.NewRequest("POST", registryURL, nil)
	req.Header.Set("X-rpc-Servers", addr)
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("rpc server: heart beat err:", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

// 心跳
// Heartbeat send a heartbeat message every once in a while
func Heartbeat(registryURL, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}

	err := sendHeartbeat(registryURL, addr)
	go func() {
		//创建一个定时器
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registryURL, addr)
		}
	}()
}
