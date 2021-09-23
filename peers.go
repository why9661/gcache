package gcache

import (
	"github.com/why9661/gcache/pb"
)

//PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	// Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}

//PeerPicker is the interface that must be implemented to locate the peer that owns a specific key.
type PeerPicker interface {
	// Pickpeer(key string) (PeerGetter, bool)
	PickPeer(key string) (PeerGetter, bool)
}
