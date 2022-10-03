package main

import (
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/tunnel"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"math/rand"
	"time"
)

func init() {
	rootCmd.AddCommand(newRun().cmd)
}

type run struct {
	cmd     *cobra.Command
	loopers int
}

func newRun() *run {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start a loop agent",
		Args:  cobra.ExactArgs(0),
	}
	r := &run{cmd: cmd}
	cmd.Run = r.run
	cmd.Flags().IntVarP(&r.loopers, "loopers", "l", 1, "Number of current loopers to start")
	return r
}

func (r *run) run(_ *cobra.Command, _ []string) {
	var loopers []*looper
	for i := 0; i < r.loopers; i++ {
		l := newLooper(i)
		loopers = append(loopers, l)
		go l.run()
	}
	for _, l := range loopers {
		<-l.done
	}
}

type looper struct {
	id   int
	done chan struct{}
}

func newLooper(id int) *looper {
	return &looper{
		id:   id,
		done: make(chan struct{}),
	}
}

func (l *looper) run() {
	logrus.Infof("starting #%d", l.id)
	defer close(l.done)
	defer logrus.Infof("stopping #%d", l.id)

	env, err := zrokdir.LoadEnvironment()
	if err != nil {
		panic(err)
	}
	zrok, err := zrokdir.ZrokClient(env.ApiEndpoint)
	if err != nil {
		panic(err)
	}
	auth := httptransport.APIKeyAuth("x-token", "header", env.ZrokToken)
	tunnelReq := tunnel.NewTunnelParams()
	tunnelReq.Body = &rest_model_zrok.TunnelRequest{
		ZitiIdentityID: env.ZitiIdentityId,
		Endpoint:       fmt.Sprintf("looper#%d", l.id),
		AuthScheme:     string(model.None),
	}
	tunnelResp, err := zrok.Tunnel.Tunnel(tunnelReq, auth)
	if err != nil {
		panic(err)
	}

	logrus.Infof("service: %v, frontend: %v", tunnelResp.Payload.Service, tunnelResp.Payload.ProxyEndpoint)
	time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)

	untunnelReq := tunnel.NewUntunnelParams()
	untunnelReq.Body = &rest_model_zrok.UntunnelRequest{
		ZitiIdentityID: env.ZitiIdentityId,
		Service:        tunnelResp.Payload.Service,
	}
	if _, err := zrok.Tunnel.Untunnel(untunnelReq, auth); err != nil {
		logrus.Errorf("error shutting down looper #%d: %v", l.id, err)
	}
}
