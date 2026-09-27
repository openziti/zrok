package limits

import (
	"testing"

	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/stretchr/testify/assert"
)

func TestNewLimitActionSimple(t *testing.T) {
	str := &store.Store{}
	zCfg := &automation.Config{}
	factory := func() (*automation.ZitiAutomation, error) { return automation.NewZitiAutomation(zCfg) }

	action := newLimitAction(str, factory)

	assert.NotNil(t, action)
	assert.Equal(t, str, action.str)
	assert.NotNil(t, action.newZiti)
}

func TestLimitAction_InterfaceCompliance(t *testing.T) {
	str := &store.Store{}
	factory := func() (*automation.ZitiAutomation, error) { return nil, nil }

	action := newLimitAction(str, factory)

	// verify it implements the AccountAction interface
	var _ AccountAction = action

	// The interface is correctly implemented - no need to test the actual method call
	// which would require database setup
}
