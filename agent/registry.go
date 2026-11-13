package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/michaelquigley/df/dd"
)

// RegistryV is the version every save writes, whatever the registry holds. agents from earlier releases refuse
// any version but "2" and never replay a refused registry, so a version 3 file cannot be replayed by an agent
// that would mishandle its contents. rolling back is the manual edit in docs/current/agent-share-lifecycle.md.
const RegistryV = "3"

// registryReadableV lists the versions LoadRegistry accepts. a version 2 file is read as it is; no conversion
// exists, and the next save writes RegistryV.
var registryReadableV = []string{"2", RegistryV}

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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var values map[string]any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("unable to parse registry '%v': %w", path, err)
	}
	// the v2.0.8 writer emitted null for unused request slices. encoding/json accepted these as zero values;
	// omit null object fields before binding a fresh registry so dd retains the same behavior.
	omitRegistryNullFields(values)
	r := &Registry{}
	if err := dd.Bind(r, values, registryBindingOptions()); err != nil {
		return nil, err
	}

	if !slices.Contains(registryReadableV, r.V) {
		return nil, fmt.Errorf("invalid registry version '%v'; expected one of '%v'", r.V, strings.Join(registryReadableV, "', '"))
	}

	return r, nil
}

func omitRegistryNullFields(value any) {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if child == nil {
				delete(value, key)
			} else {
				omitRegistryNullFields(child)
			}
		}
	case []any:
		for _, child := range value {
			omitRegistryNullFields(child)
		}
	}
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
