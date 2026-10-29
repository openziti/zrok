package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/michaelquigley/df/dd"
)

const RegistryV = "2"

type Registry struct {
	V               string                       `dd:"v"`
	PrivateAccesses []*AccessRegistryEntry       `dd:"private_accesses"`
	PublicShares    []*PublicShareRegistryEntry  `dd:"public_shares"`
	PrivateShares   []*PrivateShareRegistryEntry `dd:"private_shares"`
}

type AccessRegistryEntry struct {
	Request *AccessPrivateRequest `dd:"request"`
	Failure *FailureEntry         `dd:"failure"`
}

type PublicShareRegistryEntry struct {
	Request *SharePublicRequest `dd:"request"`
	Failure *FailureEntry       `dd:"failure"`
}

type PrivateShareRegistryEntry struct {
	Request *SharePrivateRequest `dd:"request"`
	Failure *FailureEntry        `dd:"failure"`
}

type FailureEntry struct {
	Count     int       `dd:"count"`
	LastError string    `dd:"last_error"`
	NextRetry time.Time `dd:"next_retry"`
}

func LoadRegistry(path string) (*Registry, error) {
	r := &Registry{}
	if err := dd.BindFromJSON(r, path, registryBindingOptions()); err != nil {
		return nil, err
	}

	if r.V != RegistryV {
		return nil, fmt.Errorf("invalid registry version '%v'; expected '%v'", r.V, RegistryV)
	}

	return r, nil
}

func (r *Registry) Save(path string) error {
	r.V = RegistryV
	values, err := dd.Unbind(r, registryBindingOptions())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".agent-registry-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Chmod(0644); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// registryTimestamp retains the registry v2 RFC3339 timestamp representation through dd.
type registryTimestamp struct{}

func (registryTimestamp) FromRaw(raw any) (any, error) {
	value, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("invalid registry retry timestamp")
	}
	return time.Parse(time.RFC3339Nano, value)
}
func (registryTimestamp) ToRaw(value any) (any, error) {
	stamp, ok := value.(time.Time)
	if !ok {
		return nil, fmt.Errorf("invalid registry retry timestamp")
	}
	return stamp.Format(time.RFC3339Nano), nil
}
func registryBindingOptions() *dd.Options {
	return &dd.Options{Converters: map[reflect.Type]dd.Converter{reflect.TypeOf(time.Time{}): registryTimestamp{}}}
}
