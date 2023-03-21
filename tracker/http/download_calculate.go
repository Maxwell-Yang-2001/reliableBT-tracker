package http

import (
	"net"
)

type downloadUploadParams struct {
	downloadbytes int
	uploadbytes int
}


func (t *HTTPTracker) calculate_speed(conn net.Conn, vals downloadUploadParams) {
	if vals.downloadbytes > 20 {
		t.clientError(conn, "Invalid bytes")
		return
	}
}