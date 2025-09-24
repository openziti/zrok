package agent

import (
	"encoding/json"
	"fmt"
	"os"
)

const RegistryV = "2"

type Registry struct {
	V               string                  `json:"v"`
	PrivateAccesses []*AccessPrivateRequest `json:"private_accesses,omitempty"`
	PublicShares    []*SharePublicRequest   `json:"public_shares,omitempty"`
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
	// handle migration from v1 to v2
	if r.V == "1" {
		r.V = RegistryV
	} else if r.V != RegistryV {
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
