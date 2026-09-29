package controller

import (
	"context"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
)

// mappingPublisher sends frontend mapping updates to the dynamic proxy frontends.
type mappingPublisher interface {
	Publish(ctx context.Context, frontendToken string, m dynamicProxyController.Mapping) error
}

// mappingPub is assigned from the dynamic proxy controller at startup when one is configured; nil
// otherwise, and then updates are not sent.
var mappingPub mappingPublisher

// pendingMappingUpdate is a frontend mapping change made inside a transaction. it is published only
// after that transaction commits.
type pendingMappingUpdate struct {
	frontendToken string
	mapping       dynamicProxyController.Mapping
}

// publishMappingUpdates is called only after trx.Commit() returned nil. a failed publish is logged and
// not returned: the committed rows are the truth, and a lost update is left to the frontends'
// reconciliation against them.
func publishMappingUpdates(updates []pendingMappingUpdate) {
	if mappingPub == nil {
		return
	}
	for _, update := range updates {
		if err := mappingPub.Publish(context.Background(), update.frontendToken, update.mapping); err != nil {
			dl.Errorf("error publishing mapping update '%+v' to frontend '%v': %v", update.mapping, update.frontendToken, err)
		}
	}
}
