package zrokdir

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
)

const V = "v0.3"

var migratedToXDG = false

type Metadata struct {
	V   string `json:"v"`
	Xdg bool   `json:"xdg"`
}

func checkMetadata() error {
	mf, err := metadataFile()
	fmt.Println("Checking metadata!")
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
	migratedToXDG = m.Xdg
	if !m.Xdg {
		//Check if there was a previous install. Migrate if so and mark as xdg enabled.
		fmt.Println("Should migrate to xdg...")
		if err := migrate(); err != nil {
			return errors.Wrap(err, "Unable to migrate to XDG config spec")
		}
		m.Xdg = true
		migratedToXDG = true
		//return errors.Errorf("Need to migrate to xdg")
	}
	return nil
}

func writeMetadata() error {
	mf, err := metadataFile()
	if err != nil {
		return err
	}
	data, err := json.Marshal(&Metadata{V: V, Xdg: migratedToXDG})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(mf), os.FileMode(0700)); err != nil {
		return err
	}
	if err := os.WriteFile(mf, data, os.FileMode(0600)); err != nil {
		return err
	}
	return nil
}
