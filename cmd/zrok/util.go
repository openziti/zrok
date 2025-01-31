package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/pkg/errors"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func mustGetAdminAuth() runtime.ClientAuthInfoWriter {
	adminToken := os.Getenv("ZROK_ADMIN_TOKEN")
	if adminToken == "" {
		panic("please set ZROK_ADMIN_TOKEN to a valid admin token for your zrok instance")
	}
	return httptransport.APIKeyAuth("X-TOKEN", "header", adminToken)
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
