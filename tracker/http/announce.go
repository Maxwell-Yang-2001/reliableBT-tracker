package http

import (
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/crimist/trakx/tracker/shared"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/storage"
)

type announceParams struct {
	compact  bool
	nopeerid bool
	noneleft bool
	event    string
	port     string
	hash     string
	peerid   string
	numwant  string
}

func (t *HTTPTracker) announce(conn net.Conn, vals *announceParams, ip storage.PeerIP) {
	storage.Expvar.Announces.Add(1)

	// get vars
	var hash storage.Hash
	var peerid storage.PeerID

	// hash
	if len(vals.hash) != 20 {
		t.clientError(conn, "Invalid infohash")
		return
	}
	copy(hash[:], vals.hash)

	// peerid
	if len(vals.peerid) != 20 {
		t.clientError(conn, "Invalid peerid")
		return
	}
	copy(peerid[:], vals.peerid)

	// get if stop before continuing
	if vals.event == "stopped" {
		t.peerdb.Drop(hash, peerid)
		storage.Expvar.AnnouncesOK.Add(1)
		conn.Write(shared.StringToBytes(httpSuccess))
		return
	}

	// port
	portInt, err := strconv.Atoi(vals.port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		t.clientError(conn, "Invalid port")
		return
	}

	// numwant
	numwant := int(t.conf.Tracker.Numwant.Default)

	if vals.numwant != "" {
		numwantInt, err := strconv.Atoi(vals.numwant)
		if err != nil || numwantInt < 0 {
			t.clientError(conn, "Invalid numwant")
			return
		}

		// if numwant is within our limit than listen to the client
		if numwantInt <= int(t.conf.Tracker.Numwant.Limit) {
			numwant = numwantInt
		} else {
			numwant = int(t.conf.Tracker.Numwant.Limit)
		}
	}

	peer := storage.PeerChan.Get()
	peer.Port = uint16(portInt)
	peer.IP = ip
	peer.LastSeen = time.Now().Unix()
	if vals.event == "completed" || vals.noneleft {
		peer.Complete = true
	} else {
		peer.Complete = false
	}

	t.peerdb.Save(peer, hash, peerid)
	complete, incomplete := t.peerdb.HashStats(hash)

	d := bencoding.NewDict()
	d.Int64("interval", int64(t.conf.Tracker.Announce+rand.Int31n(t.conf.Tracker.AnnounceFuzz)))
	d.Int64("complete", int64(complete))
	d.Int64("incomplete", int64(incomplete))
	if vals.compact {
		peerlist := t.peerdb.PeerListBytes(hash, numwant)
		d.StringBytes("peers", peerlist.Data)
		peerlist.Put()
	} else {
		// Escapes to heap but isn't used in prod much
		d.Any("peers", t.peerdb.PeerList(hash, numwant, vals.nopeerid))
	}

	storage.Expvar.AnnouncesOK.Add(1)
	conn.Write(shared.StringToBytes(httpSuccess))
	conn.Write(d.GetBytes())
}
