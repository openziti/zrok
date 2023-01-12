package endpoints

import "time"

type RequestHandler interface {
	Requests() func() int32
}

type Request struct {
	Stamp      time.Time
	RemoteAddr string
	Method     string
	Path       string
}
