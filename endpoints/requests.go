package endpoints

import "time"

type Request struct {
	Stamp      time.Time
	RemoteAddr string
	Method     string
	Path       string
}
