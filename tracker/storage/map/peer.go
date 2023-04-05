package gomap

import (
	"net/netip"
	"time"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/storage"
)

func (memoryDb *Memory) Save(ip netip.Addr, port uint16, complete bool, hash storage.Hash, id storage.PeerID, uploaded int64, downloaded int64, baselineProvider bool) (isBad bool) {
	// get/create the map
	memoryDb.mutex.RLock()
	peermap, ok := memoryDb.hashmap[hash]
	memoryDb.mutex.RUnlock()

	// if submap doesn't exist create it
	if !ok {
		memoryDb.mutex.Lock()
		peermap = memoryDb.makePeermap(hash)
		memoryDb.mutex.Unlock()
	}

	// if saving a baseline provider
	if baselineProvider {
		if !complete {
			// If incomplete, it will only act as a leecher in the first, thus not truly a "baseline provider" yet
			// As we want to avoid promoting it to other members, no change to the storage (neither peers nor baseline providers)
			// However, it will always be considered as a good actor thus maximum allowed peer list will be provided
			return true
		}
		peermap.mutex.RLock()
		bp, bpExists := peermap.BaselineProviders[id]
		peermap.mutex.RUnlock()

		peermap.mutex.Lock()
		// if baseline provider does not exist then create
		if !bpExists {
			bp = pools.Peers.Get()
			peermap.BaselineProviders[id] = bp
		}
		peermap.mutex.Unlock()

		// update metrics
		if !fast && !bpExists {
			stats.IPStats.Lock()
			stats.IPStats.Inc(ip)
			stats.IPStats.Unlock()
		}

		bp.LastSeen = time.Now().Unix()
		// baseline provider is never a bad actor
		return true
	}

	// get peer
	peermap.mutex.RLock()
	peer, peerExists := peermap.Peers[id]
	peermap.mutex.RUnlock()

	peermap.mutex.Lock()
	// if peer does not exist then create
	if !peerExists {
		peer = pools.Peers.Get()
		peermap.Peers[id] = peer
	}

	// update peermap completion counts
	// raw increment is 19x faster than atomic so we might as well just wrap it in the mutex
	if peerExists {
		if !peer.Complete && complete {
			peermap.Incomplete--
			peermap.Complete++
		} else if peer.Complete && !complete {
			peermap.Complete--
			peermap.Incomplete++
		}
	} else {
		if complete {
			peermap.Complete++
		} else {
			peermap.Incomplete++
		}
	}
	peermap.mutex.Unlock()

	// update metrics
	if !fast {
		if peerExists {
			// They completed
			if !peer.Complete && complete {
				stats.Leeches.Add(-1)
				stats.Seeds.Add(1)
			} else if peer.Complete && !complete { // They uncompleted?
				stats.Seeds.Add(-1)
				stats.Leeches.Add(1)
			}
			// IP changed
			if peer.IP != ip {
				stats.IPStats.Lock()
				stats.IPStats.Remove(peer.IP)
				stats.IPStats.Inc(ip)
				stats.IPStats.Unlock()
			}
		} else {
			stats.IPStats.Lock()
			stats.IPStats.Inc(ip)
			stats.IPStats.Unlock()

			if complete {
				stats.Seeds.Add(1)
			} else {
				stats.Leeches.Add(1)
			}
		}
	}

	// now to identify bad actors - we will adopt a very simple design (that does not necessarily work 100%)
	// a peer is "bad" if:
	// it is not yet incomplete
	// it uploads nothing, but downloads something
	// there have been multiple leechers in the swarm since the last time this records is made
	if !complete && peer.Uploaded == uploaded && peer.Downloaded < downloaded && peer.LeechersLastTime >= config.Config.Behavior.MinLeechers {
		isBad = true
	}

	// update peer
	peer.Complete = complete
	peer.IP = ip
	peer.Port = port
	peer.LastSeen = time.Now().Unix()
	if peer.Uploaded < uploaded {
		peer.Uploaded = uploaded
	}
	if peer.Downloaded < downloaded {
		peer.Downloaded = downloaded
	}
	peer.LeechersLastTime = peermap.Incomplete
	return
}

// delete is similar to drop but doesn't lock
func (db *Memory) delete(peer *storage.Peer, peermap *PeerMap, id storage.PeerID, baselineProvider bool) {
	if baselineProvider {
		delete(peermap.BaselineProviders, id)
	} else {
		delete(peermap.Peers, id)

		if peer.Complete {
			peermap.Complete--
		} else {
			peermap.Incomplete--
		}
	}

	if !fast {
		if !baselineProvider {
			if peer.Complete {
				stats.Seeds.Add(-1)
			} else {
				stats.Leeches.Add(-1)
			}
		}

		stats.IPStats.Lock()
		stats.IPStats.Remove(peer.IP)
		stats.IPStats.Unlock()
	}

	pools.Peers.Put(peer)
}

// Drop deletes peer or baseline provider
func (db *Memory) Drop(hash storage.Hash, id storage.PeerID, baselineProvider bool) {
	// get the peermap
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.Lock()
	// get the baseline provider and remove it
	if baselineProvider {
		peer, ok := peermap.BaselineProviders[id]
		if !ok {
			peermap.mutex.Unlock()
			return
		}

		if !fast {
			stats.IPStats.Lock()
			stats.IPStats.Remove(peer.IP)
			stats.IPStats.Unlock()
		}
		peermap.mutex.Unlock()
		// free the peer back to the pool
		pools.Peers.Put(peer)
		return
	}

	// get the peer and remove it
	peer, ok := peermap.Peers[id]
	if !ok {
		peermap.mutex.Unlock()
		return
	}
	delete(peermap.Peers, id)

	if peer.Complete {
		peermap.Complete--
	} else {
		peermap.Incomplete--
	}
	peermap.mutex.Unlock()

	if !fast {
		if peer.Complete {
			stats.Seeds.Add(-1)
		} else {
			stats.Leeches.Add(-1)
		}

		stats.IPStats.Lock()
		stats.IPStats.Remove(peer.IP)
		stats.IPStats.Unlock()
	}

	// free the peer back to the pool
	pools.Peers.Put(peer)
}
