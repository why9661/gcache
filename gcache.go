package gcache

import (
	"fmt"
	"github.com/why9661/gcache/pb"
	"github.com/why9661/gcache/singleflight"
	"log"
	"sync"
)

// A Getter loads data by a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	// namespace
	name      string
	mainCache cache
	getter    Getter
	peers     PeerPicker
	// To make sure that each key is only fetched once
	loader *singleflight.Group
}

var (
	m      sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	m.Lock()
	defer m.Unlock()
	groups[name] = g
	return g
}

// GetGroup returns the named group
func GetGroup(name string) *Group {
	m.RLock()
	g := groups[name]
	m.RUnlock()
	return g
}

// Get value by given key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("kei is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Cache] hit")
		return v, nil
	}

	return g.load(key)
}

// load loads key either by invoking the getter locally or by sending it to other machine
func (g *Group) load(key string) (value ByteView, err error) {
	//each key is only fetched once (either locally or remotely) regardless of the number of concurrent callers.
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peerGetter, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peerGetter, key); err == nil {
					return value, nil
				}
				log.Println("[gcache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
