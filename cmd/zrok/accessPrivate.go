package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/cmd/zrok/subordinate"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/endpoints/vpn"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/tui"
	"github.com/openziti/zrok/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	accessCmd.AddCommand(newAccessPrivateCommand().cmd)
}

type accessPrivateCommand struct {
	bindAddress     string
	autoMode        bool
	autoAddress     string
	autoStartPort   uint16
	autoEndPort     uint16
	headless        bool
	subordinate     bool
	forceLocal      bool
	forceAgent      bool
	responseHeaders []string
	cmd             *cobra.Command
}

func newAccessPrivateCommand() *accessPrivateCommand {
	cmd := &cobra.Command{
		Use:   "private <shareToken>",
		Short: "Create a private frontend to access a share",
		Args:  cobra.ExactArgs(1),
	}
	command := &accessPrivateCommand{cmd: cmd}
	headless := false
	if root, err := environment.LoadRoot(); err == nil {
		headless, _ = root.Headless()
	}
	cmd.Flags().BoolVar(&command.headless, "headless", headless, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable subordinate mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().BoolVar(&command.forceLocal, "force-local", false, "Skip agent detection and force local mode")
	cmd.Flags().BoolVar(&command.forceAgent, "force-agent", false, "Skip agent detection and force agent mode")
	cmd.MarkFlagsMutuallyExclusive("force-local", "force-agent")
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "127.0.0.1:9191", "The address to bind the private frontend (ignored when using '--auto')")
	cmd.Flags().BoolVar(&command.autoMode, "auto", false, "Enable automatic port detection")
	cmd.Flags().StringVar(&command.autoAddress, "auto-address", "127.0.0.1", "The address to use for automatic port detection")
	cmd.Flags().Uint16Var(&command.autoStartPort, "auto-start-port", 8080, "The starting port to use for automatic port detection")
	cmd.Flags().Uint16Var(&command.autoEndPort, "auto-end-port", 8888, "The ending port to use for automatic port detection")
	cmd.Flags().StringArrayVar(&command.responseHeaders, "response-header", []string{}, "Add a response header ('key:value')")
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	if cmd.subordinate {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	}

	root, err := environment.LoadRoot()
	if err != nil {
		cmd.error(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	if cmd.subordinate || cmd.forceLocal {
		cmd.accessLocal(args, root)
	} else {
		agent := cmd.forceAgent
		if !cmd.forceAgent {
			agent, err = agentClient.IsAgentRunning(root)
			if err != nil {
				tui.Error("error checking if agent is running", err)
			}
		}
		if agent {
			cmd.accessAgent(args, root)
		} else {
			cmd.accessLocal(args, root)
		}
	}
}

func (cmd *accessPrivateCommand) accessLocal(args []string, root env_core.Root) {
	shrToken := args[0]

	zrok, err := root.Client()
	if err != nil {
		cmd.error(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().AccountToken)
	req := share.NewAccessParams()
	req.Body.ShareToken = shrToken
	req.Body.EnvZID = root.Environment().ZitiIdentity

	accessResp, err := zrok.Share.Access(req, auth)
	if err != nil {
		cmd.error(err)
	}

	bindAddress := cmd.bindAddress
	if cmd.autoMode {
		if accessResp.Payload.BackendMode == "udpTunnel" {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(errors.New("auto-addressing is not compatible with the 'udpTunnel' backend mode"))
		}
		autoAddress, err := util.AutoListenerAddress("tcp", cmd.autoAddress, cmd.autoStartPort, cmd.autoEndPort)
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		bindAddress = autoAddress
	}

	upReq := share.NewUpdateAccessParams()
	upReq.Body.FrontendToken = accessResp.Payload.FrontendToken
	upReq.Body.BindAddress = bindAddress
	_, err = zrok.Share.UpdateAccess(upReq, auth)
	if err != nil {
		cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
		cmd.error(err)
	}

	protocol := "http://"
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		protocol = "tcp://"
	case "udpTunnel":
		protocol = "udp://"
	}

	endpointUrl, err := url.Parse(protocol + bindAddress)
	if err != nil {
		cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
		cmd.error(err)
	}

	requests := make(chan *endpoints.Request, 1024)
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
			BindAddress:  bindAddress,
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
				cmd.error(err)
			}
		}()

	case "udpTunnel":
		fe, err := udpTunnel.NewFrontend(&udpTunnel.FrontendConfig{
			BindAddress:  cmd.bindAddress,
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
			IdleTime:     time.Minute,
		})
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
				cmd.error(err)
			}
		}()

	case "socks":
		fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
			BindAddress:  bindAddress,
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
				cmd.error(err)
			}
		}()

	case "vpn":
		endpointUrl = &url.URL{
			Scheme: "VPN",
		}
		fe, err := vpn.NewFrontend(&vpn.FrontendConfig{
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
				cmd.error(err)
			}
		}()

	default:
		cfg := proxy.DefaultFrontendConfig(root.EnvironmentIdentityName())
		cfg.ShrToken = shrToken
		cfg.Address = bindAddress
		cfg.ResponseHeaders = cmd.responseHeaders
		cfg.RequestsChan = requests
		fe, err := proxy.NewFrontend(cfg)
		if err != nil {
			cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
			cmd.error(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
				cmd.error(err)
			}
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		<-c
		cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
		os.Exit(0)
	}()

	if cmd.subordinate {
		data := make(map[string]interface{})
		data[subordinate.MessageKey] = subordinate.BootMessage
		data["frontend_token"] = accessResp.Payload.FrontendToken
		data["bind_address"] = bindAddress
		jsonData, err := json.Marshal(data)
		if err != nil {
			subordinateError(err)
		}
		fmt.Println(string(jsonData))
	}

	if cmd.headless {
		logrus.Infof("access the zrok share at the following endpoint: %v", endpointUrl.String())
		for {
			select {
			case req := <-requests:
				logrus.Infof("%v -> %v %v", req.RemoteAddr, req.Method, req.Path)
			}
		}
	} else if cmd.subordinate {
		for {
			select {
			case req := <-requests:
				data := make(map[string]interface{})
				data[subordinate.MessageKey] = "access"
				data["remote-address"] = req.RemoteAddr
				data["method"] = req.Method
				data["path"] = req.Path
				jsonData, err := json.Marshal(data)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(jsonData))
			}
		}
	} else {
		mdl := newAccessModel(shrToken, endpointUrl.String())
		logrus.SetOutput(mdl)
		prg := tea.NewProgram(mdl, tea.WithAltScreen())
		mdl.prg = prg

		go func() {
			for {
				select {
				case req := <-requests:
					if req != nil {
						prg.Send(req)
					}
				}
			}
		}()

		if _, err := prg.Run(); err != nil {
			tui.Error("An error occurred", err)
		}

		close(requests)
		cmd.shutdown(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
	}
}

func (cmd *accessPrivateCommand) error(err error) {
	if cmd.subordinate {
		subordinateError(err)
	}
	if !panicInstead {
		tui.Error("unable to create private access", err)
	}
	panic(err)
}

func (cmd *accessPrivateCommand) shutdown(frontendToken, envZId, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Infof("shutting down '%v'", shrToken)
	req := share.NewUnaccessParams()
	req.Body.FrontendToken = frontendToken
	req.Body.ShareToken = shrToken
	req.Body.EnvZID = envZId
	if _, err := zrok.Share.Unaccess(req, auth); err == nil {
		logrus.Debugf("shutdown complete")
	} else {
		logrus.Errorf("error shutting down: %v", err)
	}
}

func (cmd *accessPrivateCommand) accessAgent(args []string, root env_core.Root) {
	client, conn, err := agentClient.NewClient(root)
	if err != nil {
		tui.Error("error connecting to agent", err)
	}
	defer func() { _ = conn.Close() }()

	req := &agentGrpc.AccessPrivateRequest{
		Token:           args[0],
		BindAddress:     cmd.bindAddress,
		ResponseHeaders: cmd.responseHeaders,
	}
	if cmd.autoMode {
		req.AutoMode = true
		req.AutoAddress = cmd.autoAddress
		req.AutoStartPort = uint32(cmd.autoStartPort)
		req.AutoEndPort = uint32(cmd.autoEndPort)
	}

	acc, err := client.AccessPrivate(context.Background(), req)
	if err != nil {
		tui.Error("error creating access", err)
	}

	fmt.Println(acc)
}
