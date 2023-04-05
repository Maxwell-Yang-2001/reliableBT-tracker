package udp

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) announce(announce *protocol.Announce, remote *net.UDPAddr, addrPort netip.AddrPort) {
	stats.Announces.Add(1)

	if announce.Port == 0 {
		msg := u.newClientError("bad port", announce.TransactionID, cerrFields{"addrPort": addrPort, "port": announce.Port})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if announce.NumWant < 1 {
		announce.NumWant = int32(config.Config.Numwant.Default)
	} else if announce.NumWant > int32(config.Config.Numwant.Limit) {
		announce.NumWant = int32(config.Config.Numwant.Limit)
	}

	if announce.Event == protocol.EventStopped {
		u.peerdb.Drop(announce.InfoHash, announce.PeerID, false)

		resp := protocol.AnnounceResp{
			Action:        protocol.ActionAnnounce,
			TransactionID: announce.TransactionID,
			Interval:      -1,
			Leechers:      -1,
			Seeders:       -1,
			Peers:         []byte{},
		}
		respBytes, err := resp.Marshall()
		if err != nil {
			msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		u.sock.WriteToUDP(respBytes, remote)
		return
	}

	peerComplete := false
	if announce.Event == protocol.EventCompleted || announce.Left == 0 {
		peerComplete = true
	}

	u.peerdb.Save(addrPort.Addr(), announce.Port, peerComplete, announce.InfoHash, announce.PeerID, announce.Uploaded, announce.Downloaded, false)

	complete, incomplete := u.peerdb.HashStats(announce.InfoHash)
	peers4, peers6 := u.peerdb.PeerListBytes(announce.InfoHash, uint(announce.NumWant))
	interval := int32(config.Config.Announce.Base.Seconds())
	if int32(config.Config.Announce.Fuzz.Seconds()) > 0 {
		interval += rand.Int31n(int32(config.Config.Announce.Fuzz.Seconds()))
	}

	resp := protocol.AnnounceResp{
		Action:        protocol.ActionAnnounce,
		TransactionID: announce.TransactionID,
		Interval:      interval,
		Leechers:      int32(incomplete),
		Seeders:       int32(complete),
	}

	if addrPort.Addr().Is4() {
		resp.Peers = peers4
	} else {
		resp.Peers = peers6
	}

	respBytes, err := resp.Marshall()
	pools.Peerlists4.Put(peers4)
	pools.Peerlists6.Put(peers6)

	if err != nil {
		msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	u.sock.WriteToUDP(respBytes, remote)
}
