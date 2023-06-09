package storage

import "net/netip"

type (
	// Hash stores a BitTorrent infohash.
	Hash [20]byte
	// PeerID stores a BitTorrent peer ID.
	PeerID [20]byte

	// Peer contains requied peer information for database.
	Peer struct {
		Complete         bool
		IP               netip.Addr
		Port             uint16
		LastSeen         int64
		Uploaded         int64
		Downloaded       int64
		LeechersLastTime uint16
	}

	// Reliable source contains IP and port of a known reliable source.
	ReliableSource struct {
		IP   netip.Addr
		Port uint16
	}
)
