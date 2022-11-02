package udp

import (
	"net"

	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) scrape(scrape *protocol.Scrape, remote *net.UDPAddr) {
	stats.Scrapes.Add(1)

	if len(scrape.InfoHashes) > 74 {
		msg := u.newClientError("74 hashes max", scrape.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	resp := protocol.ScrapeResp{
		Action:        protocol.ActionScrape,
		TransactionID: scrape.TransactionID,
	}

	for _, hash := range scrape.InfoHashes {
		if len(hash) != 20 {
			msg := u.newClientError("bad hash", scrape.TransactionID)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		complete, incomplete := u.peerdb.HashStats(hash)
		info := protocol.ScrapeInfo{
			Complete:   int32(complete),
			Incomplete: int32(incomplete),
			Downloaded: -1,
		}
		resp.Info = append(resp.Info, info)
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		msg := u.newServerError("ScrapeResp.Marshall()", err, scrape.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	u.sock.WriteToUDP(respBytes, remote)
}
