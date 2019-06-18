package tracker

import (
	"expvar"
	"net/http"
	"time"
)

func getInfo() (int64, int64, int64, int64, int64) {
	var peers int64
	var seeds int64
	var leeches int64
	ipmap := make(map[string]bool)

	for _, peermap := range db {
		peers += int64(len(peermap))

		for _, peer := range peermap {
			ipmap[peer.IP] = true
			if peer.Complete == true {
				seeds++
			} else {
				leeches++
			}
		}
	}

	hashes := int64(len(db))
	ips := int64(len(ipmap))

	return peers, hashes, ips, seeds, leeches
}

var (
	expvarCleanedPeers  int64
	expvarCleanedHashes int64
	expvarHits          int64
	expvarErrs          int64
)

// Expvar is for netdata
func Expvar() {
	var hitsAvg [60 * 5]int
	var hitsOld int64
	var errsOld int64

	uniqueIP := expvar.NewInt("tracker.ips")
	uniqueHash := expvar.NewInt("tracker.hashes")
	uniquePeer := expvar.NewInt("tracker.peers")
	cleanedPeers := expvar.NewInt("tracker.cleaned.peers")
	cleanedHashes := expvar.NewInt("tracker.cleaned.hashes")
	seeds := expvar.NewInt("tracker.seeds")
	leeches := expvar.NewInt("tracker.leeches")
	hits := expvar.NewInt("tracker.hits")
	hitsSec := expvar.NewInt("tracker.hits.sec")
	hits5min := expvar.NewInt("tracker.hits.5min")
	errors := expvar.NewInt("tracker.errors")
	errorsPerSec := expvar.NewInt("tracker.errorspersec")

	go http.ListenAndServe("127.0.0.1:"+trackerExpvarPort, nil) // only on localhost

	nextTime := time.Now().Truncate(time.Second)

	i := 0
	for {
		peers, hashes, ips, s, l := getInfo()
		hitsTick := expvarHits - hitsOld

		uniqueIP.Set(ips)
		uniqueHash.Set(hashes)
		uniquePeer.Set(peers)
		seeds.Set(s)
		leeches.Set(l)
		cleanedPeers.Set(expvarCleanedPeers)
		cleanedHashes.Set(expvarCleanedHashes)
		hits.Set(expvarHits)
		hitsSec.Set(hitsTick)
		hits5min.Set(func() int64 {
			hitsAvg[i] = int(hitsTick)
			i++
			if i > len(hitsAvg) {
				i = 0
			}
			total := 0

			for _, val := range hitsAvg {
				total += val
			}

			return int64(total / len(hitsAvg))
		}())
		errors.Set(expvarErrs)
		errorsPerSec.Set(expvarErrs - errsOld)

		hitsOld = expvarHits
		errsOld = expvarErrs

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
