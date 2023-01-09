package zrokdir

import (
	"encoding/json"
	"github.com/openziti-test-kitchen/zrok/tui"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

const V = "v0.3"

type Metadata struct {
	V string `json:"v"`
}

func checkMetadata() error {
	mf, err := metadataFile()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(mf)
	if err != nil {
		tui.Warning("unable to open zrokdir metadata; ignoring\n")
		return nil
	}
	m := &Metadata{}
	if err := json.Unmarshal(data, m); err != nil {
		return errors.Wrapf(err, "error unmarshaling metadata file '%v'", mf)
	}
	if m.V != V {
		return errors.Errorf("invalid zrokdir metadata version '%v'", m.V)
	}
	return nil
}

func writeMetadata() error {
	mf, err := metadataFile()
	if err != nil {
		return err
	}
	data, err := json.Marshal(&Metadata{V: V})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(mf), os.FileMode(0700)); err != nil {
		return err
	}
	if err := os.WriteFile(mf, data, os.FileMode(0400)); err != nil {
		return err
	}
	return nil
}
