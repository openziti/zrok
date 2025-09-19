package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
)

func mustGetAdminAuth() runtime.ClientAuthInfoWriter {
	adminToken := os.Getenv("ZROK_ADMIN_TOKEN")
	if adminToken == "" {
		panic("please set ZROK_ADMIN_TOKEN to a valid admin token for your zrok instance")
	}
	return httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)
}

func mustGetEnvironmentAuth() (env_core.Root, runtime.ClientAuthInfoWriter) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
	if !env.IsEnabled() {
		panic("environment is not enabled; run 'zrok enable' first")
	}
	auth := httptransport.APIKeyAuth("X-TOKEN", "header", env.Environment().AccountToken)
	return env, auth
}

func parseUrl(in string) (string, error) {
	// parse port-only urls
	if iv, err := strconv.ParseInt(in, 10, 0); err == nil {
		if iv > 0 && iv <= math.MaxUint16 {
			if iv == 443 {
				return fmt.Sprintf("https://127.0.0.1:%d", iv), nil
			}
			return fmt.Sprintf("http://127.0.0.1:%d", iv), nil
		}
		return "", errors.Errorf("ports must be between 1 and %d; %d is not", math.MaxUint16, iv)
	}

	// make sure either https:// or http:// was specified
	if !strings.HasPrefix(in, "https://") && !strings.HasPrefix(in, "http://") {
		in = "http://" + in
	}

	// parse the url
	targetEndpoint, err := url.Parse(in)
	if err != nil {
		return "", err
	}

	return targetEndpoint.String(), nil
}

func subordinateError(err error) {
	msg := make(map[string]interface{})
	msg[subordinate.MessageKey] = subordinate.ErrorMessage
	msg[subordinate.ErrorMessage] = err.Error()
	if data, err := json.Marshal(msg); err == nil {
		fmt.Println(string(data))
	} else {
		fmt.Println("{\"" + subordinate.MessageKey + "\":\"" + subordinate.ErrorMessage + "\",\"" + subordinate.ErrorMessage + "\":\"internal error\"}")
	}
	os.Exit(1)
}

// detectAndRouteToAgent handles the common pattern of checking if the agent is running
// and routing to either agent or local execution paths. This eliminates duplicate code
// found in sharePrivate, sharePublic, and accessPrivate commands.
func detectAndRouteToAgent(
	subordinate, forceLocal, forceAgent bool,
	root env_core.Root,
	localFn func(),
	agentFn func(),
) {
	// if running in subordinate mode or forced local, always use local
	if subordinate || forceLocal {
		localFn()
		return
	}

	// determine if agent is running
	agent := forceAgent
	if !forceAgent {
		var err error
		agent, err = agentClient.IsAgentRunning(root)
		if err != nil {
			tui.Error("error checking if agent is running", err)
		}
	}

	// route to appropriate handler
	if agent {
		agentFn()
	} else {
		localFn()
	}
}
