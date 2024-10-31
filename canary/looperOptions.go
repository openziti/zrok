package canary

import "time"

type LooperOptions struct {
	Iterations uint
	Timeout    time.Duration
	MinPayload uint64
	MaxPayload uint64
	MinDwell   time.Duration
	MaxDwell   time.Duration
	MinPacing  time.Duration
	MaxPacing  time.Duration
}
