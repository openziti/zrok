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

	action := newLimitAction(str, zCfg)

	assert.NotNil(t, action)
	assert.Equal(t, str, action.str)
	assert.Equal(t, zCfg, action.zCfg)
}

func TestLimitAction_InterfaceCompliance(t *testing.T) {
	str := &store.Store{}
	zCfg := &automation.Config{}

	action := newLimitAction(str, zCfg)

	// verify it implements the AccountAction interface
	var _ AccountAction = action

	// The interface is correctly implemented - no need to test the actual method call
	// which would require database setup
}