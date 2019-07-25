package tracker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Syc0x00/Trakx/bencoding"
	httptracker "github.com/Syc0x00/Trakx/tracker/http"
	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	"go.uber.org/zap"
)

// Run runs the tracker
func Run() {
	// Init shared stuff
	if err := shared.Init(); err != nil {
		panic(err)
	}

	go handleSigs()
	if shared.Config.Tracker.Ports.Expvar != 0 {
		go Expvar()
	}

	// HTTP tracker / routes
	initRoutes()

	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/stats", stats)

	if shared.Config.Tracker.HTTP {
		shared.Logger.Info("http tracker enabled")
		trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
		trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)
	} else {
		// TODO: Interval is the only thing needed on qBit but need to test other clients
		d := bencoding.NewDict()
		d.Add("interval", 432000) // 5 days
		/* Need to consider these
		d.Add("failure reason", "REMOVE THIS TRACKER")
		d.Add("complete", -1)
		d.Add("incomplete", -1)
		d.Add("peers", "")*/
		errResp := []byte(d.Get())

		trackerMux.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {})
		trackerMux.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
			w.Write(errResp)
		})
	}

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", shared.Config.Tracker.Ports.HTTP),
		Handler:      trackerMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 7 * time.Second,
		IdleTimeout:  0,
	}
	server.SetKeepAlivesEnabled(false)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			shared.Logger.Error("ListenAndServe()", zap.Error(err))
		}
	}()

	// UDP tracker
	if shared.Config.Tracker.Ports.UDP != 0 {
		shared.Logger.Info("udp tracker enabled")
		udptracker.Run(time.Duration(shared.Config.Database.Conn.Trim) * time.Second)
	}

	select {} // block forever
}
