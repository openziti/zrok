package sdk

import (
	"errors"
	"fmt"
	"github.com/openziti/zrok/environment/env_core"
	"io"
	"net/http"
)

func Overview(root env_core.Root) (string, error) {
	if !root.IsEnabled() {
		return "", errors.New("environment is not enabled; enable with 'zrok enable' first!")
	}

	client := &http.Client{}
	apiEndpoint, _ := root.ApiEndpoint()
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v1/overview", apiEndpoint), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-TOKEN", root.Environment().AccountToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	json, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()

	return string(json), nil
}
