package limits

import "github.com/openziti/zrok/controller/store"

type AccountStrategy interface {
	HandleAccount(a *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error
}

type EnvironmentStrategy interface {
	HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error
}

type ShareStrategy interface {
	HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error
}
