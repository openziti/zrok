package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go/log"
	"github.com/openziti-test-kitchen/zrok/cmd/zrok/endpoint_ui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"html/template"
	"net"
	"net/http"
	"time"
)

func init() {
	testCmd.AddCommand(newTestEndpointCommand().cmd)
}

type testEndpointCommand struct {
	address string
	port    uint16
	t       *template.Template
	cmd     *cobra.Command
}

func newTestEndpointCommand() *testEndpointCommand {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Start a simple HTTP endpoint",
		Args:  cobra.ExactArgs(0),
	}
	command := &testEndpointCommand{cmd: cmd}
	var err error
	if command.t, err = template.ParseFS(endpoint_ui.FS, "index.gohtml"); err != nil {
		if !panicInstead {
			showError("unable to parse index template", err)
		}
		panic(err)
	}
	cmd.Flags().StringVarP(&command.address, "address", "a", "127.0.0.1", "The address for the HTTP listener")
	cmd.Flags().Uint16VarP(&command.port, "port", "P", 9090, "The port for the HTTP listener")
	cmd.Run = command.run
	return command
}

func (cmd *testEndpointCommand) run(_ *cobra.Command, _ []string) {
	http.HandleFunc("/", cmd.serveIndex)
	if err := http.ListenAndServe(fmt.Sprintf("%v:%d", cmd.address, cmd.port), nil); err != nil {
		if !panicInstead {
			showError("unable to start http listener", err)
		}
		panic(err)
	}
}

func (cmd *testEndpointCommand) serveIndex(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("%v {%v} | %v -> /index.gohtml", r.RemoteAddr, r.Host, r.RequestURI)
	if err := cmd.t.Execute(w, newEndpointData(r)); err != nil {
		log.Error(err)
	}
}

type endpointData struct {
	Now        time.Time
	RemoteAddr string
	Host       string
	HostDetail string
	Ips        string
	HostHeader string
	Headers    map[string][]string
}

func newEndpointData(r *http.Request) *endpointData {
	ed := &endpointData{
		Now:        time.Now(),
		HostHeader: r.Host,
		Headers:    r.Header,
		RemoteAddr: r.RemoteAddr,
	}
	ed.getHostInfo()
	ed.getIps()
	return ed
}

func (ed *endpointData) getHostInfo() {
	host, hostDetail, err := getHost()
	if err != nil {
		logrus.Errorf("error getting host detail: %v", err)
	}
	ed.Host = host
	ed.HostDetail = hostDetail
}

func (ed *endpointData) getIps() {
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if len(ed.Ips) != 0 {
					ed.Ips += ", "
				}
				ed.Ips += ipnet.IP.String()
			}
		}
	}
}
