package gcache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/why9661/gcache/consistenthash"
	pb "github.com/why9661/gcache/pb"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const defaultBasePath = "/_gcache/"
const defaultReplicas = 50

type HTTPPool struct {
	basePath    string // default: /_cache/
	selfPath    string // eg: localhost:8000
	mu          sync.Mutex
	peers       *consistenthash.CHash
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		selfPath: self,
		basePath: defaultBasePath,
	}
}

//Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.selfPath, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(body)
}

// Set updates the pool's list of peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.peers == nil {
		p.peers = consistenthash.New(defaultReplicas, nil)
	}
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.selfPath {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

type httpGetter struct {
	// eg:http://localhost:8001/_cache/
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	// eg: http://localhost:8001/_cache/scores/why
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
