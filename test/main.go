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

func createGroup() *gcache.Group {
	return gcache.NewGroup("scores", 2<<10, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *gcache.Group) {
	peer := gcache.NewHTTPPool(addr)
	peer.Set(addrs...)
	g.RegisterPeers(peer)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peer))
}

func startAPIServer(apiAddr string, g *gcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "gcache server port")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
