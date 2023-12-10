package cache

import (
	"fmt"
	"net/http"
	"strings"
)

const DefaultBasePath = "/geecache/"

type HTTPPool struct {
	addr     string
	basePath string
}

func (pool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, pool.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	fmt.Println(r.URL.Path)
	// fmt.Println(r.URL.Path[len(pool.basePath):], len(pool.basePath))
	parts := strings.SplitN(r.URL.Path[len(pool.basePath):], "/", 2)

	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(parts[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func NewHTTPPool(addr string, basePath string) *HTTPPool {
	return &HTTPPool{
		addr:     addr,
		basePath: basePath,
	}
}
