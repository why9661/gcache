package main

import (
	"flag"
	"fmt"
	"github.com/why9661/gcache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Mark": "566",
	"Jack": "589",
	"Why":  "888",
}

func startCacheServer(addr string, addrs []string, gee *gocache.Group) {
	peers := gocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:8004"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := gocache.NewGroup("scores", 1<<10, gocache.GetterFunc(func(k string) ([]byte, error) {
		log.Println("[SlowDB] search key", k)
		if v, ok := db[k]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", k)
	}))

	if api {
		go startAPIServer(apiAddr, group)
	}

	startCacheServer(addrMap[port], []string(addrs), group)

}
