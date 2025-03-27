package agent

import (
	"encoding/json"
	"os"
)

type Registry struct {
	ReservedShares  []*ShareReservedRequest
	PrivateAccesses []*AccessPrivateRequest
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
	return r, nil
}

func (r *Registry) Save(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
