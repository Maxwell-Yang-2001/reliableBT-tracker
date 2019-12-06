package gomap

import (
	"errors"
	"sync"
	"time"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	hashMapAlloc = 1000000
	peerMapAlloc = 1
)

type subPeerMap struct {
	sync.RWMutex
	peers map[storage.PeerID]*storage.Peer
}

type Memory struct {
	mu      sync.RWMutex
	hashmap map[storage.Hash]*subPeerMap

	backup storage.Backup
	conf   *shared.Config
}

func (db *Memory) Init(conf *shared.Config, backup storage.Backup) error {
	*db = Memory{
		conf:   conf,
		backup: backup,
	}

	if err := db.backup.Init(db); err != nil {
		return errors.New("Failed to initialize backup: " + err.Error())
	}
	if err := db.backup.Load(); err != nil {
		return errors.New("Failed to load backup: " + err.Error())
	}

	if conf.Database.Peer.Write > 0 {
		go shared.RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, func() {
			db.backup.Save()
		})
	}
	if conf.Database.Peer.Trim > 0 {
		go shared.RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, db.Trim)
	}

	return nil
}

func (db *Memory) make() {
	db.hashmap = make(map[storage.Hash]*subPeerMap, hashMapAlloc)
}

func (db *Memory) makePeermap(h *storage.Hash) (peermap *subPeerMap) {
	// build struct and assign
	peermap = new(subPeerMap)
	peermap.peers = make(map[storage.PeerID]*storage.Peer, peerMapAlloc)
	db.hashmap[*h] = peermap
	return
}

func (db *Memory) Backup() storage.Backup {
	return db.backup
}

func (db *Memory) Check() bool {
	return db.hashmap != nil
}

func (db *Memory) Trim() {
	start := time.Now()
	db.conf.Logger.Info("Trimming database")
	peers, hashes := db.trim()
	if peers < 1 && hashes < 1 {
		db.conf.Logger.Info("Can't trim database: database empty")
	} else {
		db.conf.Logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
	}
}

func (db *Memory) trim() (peers, hashes int) {
	now := time.Now().Unix()

	db.mu.RLock()
	for hash, peermap := range db.hashmap {
		db.mu.RUnlock()

		peermap.Lock()
		for id, peer := range peermap.peers {
			if now-peer.LastSeen > db.conf.Database.Peer.Timeout {
				db.delete(peer, peermap, &id)
				peers++
			}
		}
		peersize := len(peermap.peers)
		peermap.Unlock()

		if peersize == 0 {
			db.mu.Lock()
			delete(db.hashmap, hash)
			db.mu.Unlock()
			hashes++
		}
		db.mu.RLock()
	}
	db.mu.RUnlock()

	return
}