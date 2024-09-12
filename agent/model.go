package agent

import (
	"bytes"
	"encoding/json"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/agent/proctree"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"time"
)

type share struct {
	token                     string
	frontendEndpoints         []string
	target                    string
	basicAuth                 []string
	frontendSelection         []string
	shareMode                 sdk.ShareMode
	backendMode               sdk.BackendMode
	reserved                  bool
	insecure                  bool
	oauthProvider             string
	oauthEmailAddressPatterns []string
	oauthCheckInterval        time.Duration
	closed                    bool
	accessGrants              []string

	process    *proctree.Child
	readBuffer bytes.Buffer
	ready      chan struct{}
}

func (s *share) tail(data []byte) {
	s.readBuffer.Write(data)
	if line, err := s.readBuffer.ReadString('\n'); err == nil {
		if s.token == "" {
			in := make(map[string]interface{})
			if err := json.Unmarshal([]byte(line), &in); err == nil {
				if v, found := in["token"]; found {
					if str, ok := v.(string); ok {
						s.token = str
					}
				}
				if v, found := in["frontend_endpoints"]; found {
					if vArr, ok := v.([]interface{}); ok {
						for _, v := range vArr {
							if str, ok := v.(string); ok {
								s.frontendEndpoints = append(s.frontendEndpoints, str)
							}
						}
					}
				}
			}
			close(s.ready)
		} else {
			pfxlog.ChannelLogger(s.token).Info(string(line))
		}
	} else {
		s.readBuffer.WriteString(line)
	}
}

type access struct {
	token string

	bindAddress     string
	responseHeaders []string

	process *proctree.Child
}

type agentGrpcImpl struct {
	agentGrpc.UnimplementedAgentServer
	a *Agent
}
