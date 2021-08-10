package main

import (
	"flag"
	"fmt"
	"github.com/why9661/gcache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup(name string, cacheBytes int64, getter gcache.Getter) *gcache.Group {
	return gcache.NewGroup(name, cacheBytes, getter)
}

func startCacheServer(addr string, addrs []string, group *gcache.Group) {
	peer := gcache.NewHTTPPool(addr)
	peer.Set(addrs...)
	group.RegisterPeers(peer)
	log.Println("gcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peer))
}

func startAPIServer(addr string, group *gcache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(view.ByteSlice())
	})
	log.Println("fontend server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "gcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9000"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group_1 := createGroup("scores", 2<<10, gcache.GetterFunc(func(key string) ([]byte, error) {
		log.Printf("[Slow] search key: %s\n", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("No such key in db: %s\n", key)
	}))

	if api {
		go startAPIServer(apiAddr, group_1)
	}

	startCacheServer(addrMap[port], addrs, group_1)
}
