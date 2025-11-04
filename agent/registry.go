package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const RegistryV = "2"

type Registry struct {
	V               string                       `json:"v"`
	PrivateAccesses []*AccessRegistryEntry       `json:"private_accesses,omitempty"`
	PublicShares    []*PublicShareRegistryEntry  `json:"public_shares,omitempty"`
	PrivateShares   []*PrivateShareRegistryEntry `json:"private_shares,omitempty"`
}

type AccessRegistryEntry struct {
	Request *AccessPrivateRequest `json:"request"`
	Failure *FailureEntry         `json:"failure,omitempty"`
}

type PublicShareRegistryEntry struct {
	Request *SharePublicRequest `json:"request"`
	Failure *FailureEntry       `json:"failure,omitempty"`
}

type PrivateShareRegistryEntry struct {
	Request *SharePrivateRequest `json:"request"`
	Failure *FailureEntry        `json:"failure,omitempty"`
}

type FailureEntry struct {
	Count     int       `json:"count,omitempty"`
	LastError string    `json:"last_error,omitempty"`
	NextRetry time.Time `json:"next_retry,omitempty"`
}

func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := &Registry{}
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	}

	if r.V != RegistryV {
		return nil, fmt.Errorf("invalid registry version '%v'; expected '%v'", r.V, RegistryV)
	}

	return r, nil
}

func (r *Registry) Save(path string) error {
	r.V = RegistryV
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
