package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShareDefaultPermissionMode(t *testing.T) {
	shr := &Share{}
	assert.Equal(t, OpenPermissionMode, shr.PermissionMode)

	var shr2 Share
	assert.Equal(t, OpenPermissionMode, shr2.PermissionMode)
}
