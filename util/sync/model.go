package sync

import "time"

type Object struct {
	Path     string
	Size     int64
	Modified time.Time
	ETag     string
}

type Target interface {
	Inventory() ([]*Object, error)
}
