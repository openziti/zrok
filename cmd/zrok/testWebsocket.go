package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func init() {
	testCmd.AddCommand(newTestWebsocketCommand().cmd)
}

type testWebsocketCommand struct {
	cmd *cobra.Command

	identityJsonFile string
	serviceName      string
	enableZiti       bool
}

func newTestWebsocketCommand() *testWebsocketCommand {
	cmd := &cobra.Command{
		Use:  "websocket",
		Args: cobra.RangeArgs(0, 1),
	}

	command := &testWebsocketCommand{cmd: cmd}

	cmd.Flags().BoolVar(&command.enableZiti, "ziti", false, "Enable the usage of a ziti network")
	cmd.Flags().StringVar(&command.identityJsonFile, "ziti-identity", "", "Path to Ziti Identity json file")
	cmd.Flags().StringVar(&command.serviceName, "ziti-name", "", "Name of the Ziti Service")

	cmd.Run = command.run
	return command
}

func (cmd *testWebsocketCommand) run(_ *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*6)
	defer cancel()
	opts := &websocket.DialOptions{}
	var addr string
	if cmd.enableZiti {
		identityJsonBytes, err := os.ReadFile(cmd.identityJsonFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read identity config JSON from file %s: %s\n", cmd.identityJsonFile, err)
			os.Exit(1)
		}
		if len(identityJsonBytes) == 0 {
			fmt.Fprintf(os.Stderr, "Error: When running a ziti enabled service must have ziti identity provided\n\n")
			flag.Usage()
			os.Exit(1)
		}

		cfg := &ziti.Config{}
		err = json.Unmarshal(identityJsonBytes, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load ziti configuration JSON: %v", err)
			os.Exit(1)
		}
		zitiContext, err := ziti.NewContext(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load ziti context: %v", err)
			os.Exit(1)
		}
		dial := func(_ context.Context, _, addr string) (net.Conn, error) {
			service := strings.Split(addr, ":")[0]
			return zitiContext.Dial(service)
		}

		zitiTransport := http.DefaultTransport.(*http.Transport).Clone()
		zitiTransport.DialContext = dial

		opts.HTTPClient = &http.Client{Transport: zitiTransport}

		addr = cmd.serviceName
	} else {
		if len(args) == 0 {
			logrus.Error("address required if not using ziti")
			flag.Usage()
			os.Exit(1)
		}
		addr = args[0]
	}

	logrus.Info(fmt.Sprintf("http://%s/echo", addr))
	c, _, err := websocket.Dial(ctx, fmt.Sprintf("http://%s/echo", addr), opts)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	logrus.Info("writing to server...")
	err = wsjson.Write(ctx, c, "hi")
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("reading response...")
	typ, dat, err := c.Read(ctx)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info(typ)
	logrus.Info(string(dat))

	c.Close(websocket.StatusNormalClosure, "")
}
