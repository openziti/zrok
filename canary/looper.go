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
	totalBytes := uint64(0)
	totalXferRate := uint64(0)
	totalErrors := uint(0)
	totalMismatches := uint(0)
	totalLoops := uint(0)
	for i, result := range results {
		totalBytes += result.Bytes
		deltaSeconds := result.StopTime.Sub(result.StartTime).Seconds()
		xferRate := uint64(float64(result.Bytes) / deltaSeconds)
		totalXferRate += xferRate
		totalErrors += result.Errors
		totalMismatches += result.Mismatches
		totalLoops += result.Loops
		logrus.Infof("looper #%d: %d loops, %v, %d errors, %d mismatches, %s/sec", i, result.Loops, util.BytesToSize(int64(result.Bytes)), result.Errors, result.Mismatches, util.BytesToSize(int64(xferRate)))
	}
	logrus.Infof("total: %d loops, %v, %d errors, %d mismatches, %s/sec", totalLoops, util.BytesToSize(int64(totalBytes)), totalErrors, totalMismatches, util.BytesToSize(int64(totalXferRate)))
}
