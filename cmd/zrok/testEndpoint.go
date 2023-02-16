package main

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/cmd/zrok/endpointUi"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
	"html/template"
	"io"
	"net"
	"net/http"
	"nhooyr.io/websocket"
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
	if command.t, err = template.ParseFS(endpointUi.FS, "index.gohtml"); err != nil {
		if !panicInstead {
			tui.Error("unable to parse index template", err)
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
	http.HandleFunc("/echo", cmd.websocketEcho)
	if err := http.ListenAndServe(fmt.Sprintf("%v:%d", cmd.address, cmd.port), nil); err != nil {
		if !panicInstead {
			tui.Error("unable to start http listener", err)
		}
		panic(err)
	}
}

func (cmd *testEndpointCommand) serveIndex(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("%v {%v} | %v -> /index.gohtml", r.RemoteAddr, r.Host, r.RequestURI)
	if err := cmd.t.Execute(w, newEndpointData(r)); err != nil {
		logrus.Error(err)
	}
}

func (cmd *testEndpointCommand) websocketEcho(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer func() { _ = c.Close(websocket.StatusInternalError, "connection terminated") }()

	l := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	for {
		err = cmd.doEcho(r.Context(), c, l)
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			return
		}
		if err != nil {
			logrus.Errorf("failed to echo for '%v': %v", r.RemoteAddr, err)
			return
		}
	}
}

func (cmd *testEndpointCommand) doEcho(ctx context.Context, c *websocket.Conn, l *rate.Limiter) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	err := l.Wait(ctx)
	if err != nil {
		return err
	}

	typ, r, err := c.Reader(ctx)
	if err != nil {
		return err
	}

	w, err := c.Writer(ctx, typ)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("i received: "))
	_, err = io.Copy(w, r)
	if err != nil {
		return errors.Wrap(err, "failed to copy")
	}

	err = w.Close()
	return err
}

type endpointData struct {
	RequestedPath string
	Now           time.Time
	RemoteAddr    string
	Host          string
	HostDetail    string
	Ips           string
	HostHeader    string
	Headers       map[string][]string
}

func newEndpointData(r *http.Request) *endpointData {
	ed := &endpointData{
		RequestedPath: r.RequestURI,
		Now:           time.Now(),
		HostHeader:    r.Host,
		Headers:       r.Header,
		RemoteAddr:    r.RemoteAddr,
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
