package http

import (
	"net"
	"reflect"

	"github.com/crimist/trakx/pools"
)

type downloadUploadParams struct {
	downloadbytes int
	uploadbytes int
	infohash string
}


func (t *HTTPTracker) calculate_speed(conn net.Conn, vals downloadUploadParams) {
	if reflect.ValueOf(vals).IsZero() {
		t.clientError(conn, "no params")
		return
	}

	if vals.infohash == "" {
		t.clientError(conn, "no infohash")
		return
	} 

	if vals.downloadbytes == 0 {
		t.clientTorrentHashToDownload[vals.infohash] = 0
		t.clientError(conn, "no downloads occurring")
		return
	}

	if lastDownloadAmount, ok := t.clientTorrentHashToDownload[vals.infohash]; ok {
		t.downloadSpeed = (vals.downloadbytes-lastDownloadAmount)
		t.clientTorrentHashToDownload[vals.infohash] = vals.downloadbytes
	} else {
		t.clientTorrentHashToDownload[vals.infohash] = vals.downloadbytes
		t.downloadSpeed = vals.downloadbytes
		// download per second 
	}
	// fmt.Println("downloadspeed")
	dictionary := pools.Dictionaries.Get()
	dictionary.Int64("downloadSpeed", int64(t.downloadSpeed))
	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))
	pools.Dictionaries.Put(dictionary)
}