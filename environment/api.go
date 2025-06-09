package environment

import (
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/environment/env_v0_3"
	"github.com/openziti/zrok/environment/env_v0_4"
	"github.com/pkg/errors"
)

// SetRootDirName allows setting a custom name for the root directory.
// This should be called before any other environment operations.
func SetRootDirName(name string) {
	env_v0_4.SetRootDirName(name)
}

func LoadRoot() (env_core.Root, error) {
	if assert, err := env_v0_4.Assert(); assert && err == nil {
		return env_v0_4.Load()
	} else if assert, err := env_v0_3.Assert(); assert && err == nil {
		return env_v0_3.Load()
	} else {
		return env_v0_4.Default()
	}
}

func IsLatest(r env_core.Root) bool {
	if r == nil {
		return false
	}
	if r.Metadata() == nil {
		return false
	}
	if r.Metadata().V == env_v0_4.V {
		return true
	}
	return false
}

func UpdateRoot(r env_core.Root) (env_core.Root, error) {
	newR, err := env_v0_4.Update(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update environment")
	}
	return newR, nil
}
