package canary

import (
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"time"
)

type LooperOptions struct {
	Iterations     uint
	StatusInterval uint
	Timeout        time.Duration
	MinPayload     uint64
	MaxPayload     uint64
	MinDwell       time.Duration
	MaxDwell       time.Duration
	MinPacing      time.Duration
	MaxPacing      time.Duration
}

type LooperResults struct {
	StartTime  time.Time
	StopTime   time.Time
	Loops      uint
	Errors     uint
	Mismatches uint
	Bytes      uint64
}

func ReportLooperResults(results []*LooperResults) {
	totalXfer := uint64(0)
	totalErrors := uint(0)
	totalMismatches := uint(0)
	totalLoops := uint(0)
	for i, result := range results {
		deltaSeconds := result.StopTime.Sub(result.StartTime).Seconds()
		xfer := uint64(float64(result.Bytes) / deltaSeconds)
		totalXfer += xfer
		totalErrors += result.Errors
		totalMismatches += result.Mismatches
		xferSec := util.BytesToSize(int64(xfer))
		totalLoops += result.Loops
		logrus.Infof("looper #%d: %d loops, %d errors, %d mismatches, %s/sec", i, result.Loops, result.Errors, result.Mismatches, xferSec)
	}
	totalXferSec := util.BytesToSize(int64(totalXfer))
	logrus.Infof("total: %d loops, %d errors, %d mismatches, %s/sec", totalLoops, totalErrors, totalMismatches, totalXferSec)
}
