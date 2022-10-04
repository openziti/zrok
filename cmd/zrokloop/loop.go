package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti-test-kitchen/zrok/model"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/tunnel"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"time"
)

func init() {
	rootCmd.AddCommand(newRun().cmd)
}

type run struct {
	cmd         *cobra.Command
	loopers     int
	iterations  int
	statusEvery int
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
	cmd.Flags().IntVarP(&r.iterations, "iterations", "i", 1, "Number of iterations per looper")
	cmd.Flags().IntVarP(&r.statusEvery, "status-every", "E", 100, "Show status every # iterations")
	return r
}

func (r *run) run(_ *cobra.Command, _ []string) {
	var loopers []*looper
	for i := 0; i < r.loopers; i++ {
		l := newLooper(i, r)
		loopers = append(loopers, l)
		go l.run()
	}
	for _, l := range loopers {
		<-l.done
	}
}

type looper struct {
	id       int
	runi     *run
	done     chan struct{}
	listener edge.Listener
}

func newLooper(id int, runi *run) *looper {
	return &looper{
		id:   id,
		runi: runi,
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
	zif, err := zrokdir.ZitiIdentityFile("environment")
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

	logrus.Infof("looper #%d, service: %v, frontend: %v", l.id, tunnelResp.Payload.Service, tunnelResp.Payload.ProxyEndpoint)
	go l.serviceListener(zif, tunnelResp.Payload.Service)

	time.Sleep(1 * time.Second)

	for i := 0; i < l.runi.iterations; i++ {
		if i > 0 && i%l.runi.statusEvery == 0 {
			logrus.Infof("looper #%d: iteration #%d", l.id, i)
		}
		outpayload := make([]byte, 10240)
		outbase64 := base64.StdEncoding.EncodeToString(outpayload)
		rand.Read(outpayload)
		if req, err := http.NewRequest("POST", tunnelResp.Payload.ProxyEndpoint, bytes.NewBufferString(outbase64)); err == nil {
			client := &http.Client{Timeout: time.Second * 10}
			if resp, err := client.Do(req); err == nil {
				inpayload := new(bytes.Buffer)
				io.Copy(inpayload, resp.Body)
				inbase64 := inpayload.String()
				if inbase64 != outbase64 {
					logrus.Errorf("looper #%d payload mismatch!", l.id)
				} else {
					logrus.Debugf("looper #%d payload match", l.id)
				}
			} else {
				logrus.Errorf("looper #%d error: %v", l.id, err)
			}
		} else {
			logrus.Errorf("looper #%d error creating request: %v", l.id, err)
		}
	}
	logrus.Infof("looper #%d: complete", l.id)

	if l.listener != nil {
		if err := l.listener.Close(); err != nil {
			logrus.Errorf("looper #%d error closing listener: %v", l.id, err)
		}
	}

	untunnelReq := tunnel.NewUntunnelParams()
	untunnelReq.Body = &rest_model_zrok.UntunnelRequest{
		ZitiIdentityID: env.ZitiIdentityId,
		Service:        tunnelResp.Payload.Service,
	}
	if _, err := zrok.Tunnel.Untunnel(untunnelReq, auth); err != nil {
		logrus.Errorf("error shutting down looper #%d: %v", l.id, err)
	}
}

func (l *looper) serviceListener(zitiIdPath string, svcId string) {
	zcfg, err := config.NewFromFile(zitiIdPath)
	if err != nil {
		logrus.Errorf("error opening ziti config '%v': %v", zitiIdPath, err)
		return
	}
	opts := ziti.ListenOptions{
		ConnectTimeout: 5 * time.Minute,
		MaxConnections: 10,
	}
	if l.listener, err = ziti.NewContextWithConfig(zcfg).ListenWithOptions(svcId, &opts); err == nil {
		if err := http.Serve(l.listener, l); err != nil {
			logrus.Errorf("looper #%d, error serving: %v", l.id, err)
		}
	} else {
		logrus.Errorf("looper #%d, error listening: %v", l.id, err)
	}
}

func (l *looper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	io.Copy(buf, r.Body)
	w.Write(buf.Bytes())
}
