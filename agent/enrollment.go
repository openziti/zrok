package agent

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

const EnrollmentV = "1"

type Enrollment struct {
	V     string `json:"v"`
	Token string `json:"token"`
}

func NewEnrollment(token string) *Enrollment {
	return &Enrollment{Token: token}
}

func LoadEnrollment(path string) (*Enrollment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	e := &Enrollment{}
	if err := json.Unmarshal(data, e); err != nil {
		return nil, err
	}
	if e.V != EnrollmentV {
		return nil, errors.Errorf("invalid enrollment version '%v'; expected '%v'", e.V, EnrollmentV)
	}
	return e, nil
}

func (e *Enrollment) Save(path string) error {
	e.V = EnrollmentV
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
