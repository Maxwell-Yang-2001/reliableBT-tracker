package gomap

import (
	"encoding/binary"
	"errors"
	"math/rand"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/storage"
)

// Hashes gets the number of hashes
func (db *Memory) Hashes() int {
	// race condition but doesn't matter as it's just for metrics
	return len(db.hashmap)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *Memory) HashStats(hash storage.Hash) (complete, incomplete uint16) {
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()
	complete = peermap.Complete
	incomplete = peermap.Incomplete
	peermap.mutex.RUnlock()

	return
}

// Obtain one random baseline provider from the available ones in the swarm if possible
func (db *Memory) BaselineProvider(hash storage.Hash, compact bool, removePeerId bool) (baselineProvider []byte, err error) {
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return baselineProvider, errors.New("No peermap at specified hash found")
	}

	peermap.mutex.RLock()

	numBaselineProviders := len(peermap.BaselineProviders)
	if numBaselineProviders == 0 {
		peermap.mutex.RUnlock()
		return baselineProvider, errors.New("No baseline provider for specified hash found")
	}

	// naive: Randomly choose a baseline provider and return it
	randIndex := rand.Intn(numBaselineProviders)
	for id, knownProvider := range peermap.BaselineProviders {
		if randIndex == 0 {
			if compact {
				// if compact, put each byte based on IPV4/V6
				if knownProvider.IP.Is6() {
					baselineProvider = make([]byte, 18)
					copy(baselineProvider[:16], knownProvider.IP.AsSlice())
					binary.BigEndian.PutUint16(baselineProvider[16:18], knownProvider.Port)
					baselineProvider = baselineProvider[:18]
				} else {
					baselineProvider = make([]byte, 6)
					copy(baselineProvider[:4], knownProvider.IP.AsSlice())
					binary.BigEndian.PutUint16(baselineProvider[4:6], knownProvider.Port)
				}
			} else {
				// otherwise, use a dictionary and place necessary information
				dictionary := pools.Dictionaries.Get()
				if !removePeerId {
					dictionary.String("peer id", string(id[:]))
				}
				dictionary.String("ip", knownProvider.IP.String())
				dictionary.Int64("port", int64(knownProvider.Port))

				dictBytes := dictionary.GetBytes()
				baselineProvider := make([]byte, len(dictBytes))
				copy(baselineProvider, dictBytes)

				dictionary.Reset()
				pools.Dictionaries.Put(dictionary)
			}

			peermap.mutex.RUnlock()
			return baselineProvider, nil
		}

		randIndex -= 1
	}

	peermap.mutex.RUnlock()
	return baselineProvider, errors.New("Baseline provider picking goes wrong")
}

// PeerList returns a peer list for the given hash capped at max
func (db *Memory) PeerList(hash storage.Hash, numWant uint, removePeerId bool) (peers [][]byte) {
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()

	if numPeers := uint(len(peermap.Peers)); numWant > numPeers {
		numWant = numPeers
	}

	if numWant == 0 {
		peermap.mutex.RUnlock()
		return
	}

	var i uint
	peers = make([][]byte, numWant)
	dictionary := pools.Dictionaries.Get()

	for id, peer := range peermap.Peers {
		if !removePeerId {
			dictionary.String("peer id", string(id[:]))
		}
		dictionary.String("ip", peer.IP.String())
		dictionary.Int64("port", int64(peer.Port))

		dictBytes := dictionary.GetBytes()
		peers[i] = make([]byte, len(dictBytes))
		copy(peers[i], dictBytes)

		dictionary.Reset()

		i++
		if i == numWant {
			break
		}
	}

	peermap.mutex.RUnlock()
	pools.Dictionaries.Put(dictionary)

	return
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *Memory) PeerListBytes(hash storage.Hash, numWant uint) (peers4 []byte, peers6 []byte) {
	peers4 = pools.Peerlists4.Get()
	peers6 = pools.Peerlists6.Get()

	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()
	if numPeers := uint(len(peermap.Peers)); numWant > numPeers {
		numWant = numPeers
	}

	if numWant == 0 {
		peermap.mutex.RUnlock()
		return
	}

	var pos4, pos6 int
	for _, peer := range peermap.Peers {
		if peer.IP.Is6() {
			copy(peers6[pos6:pos6+16], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers6[pos6+16:pos6+18], peer.Port)
			pos6 += 18
			if pos6+18 > cap(peers6) {
				break
			}
		} else {
			copy(peers4[pos4:pos4+4], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers4[pos4+4:pos4+6], peer.Port)
			pos4 += 6
			if pos4+6 > cap(peers4) {
				break
			}
		}
	}
	peermap.mutex.RUnlock()

	peers4 = peers4[:pos4]
	peers6 = peers6[:pos6]

	return
}
