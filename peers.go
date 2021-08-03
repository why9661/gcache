package gcache

//PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

//PeerPicker is the interface that must be implemented to locate the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}
