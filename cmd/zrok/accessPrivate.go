package main

import (
	"context"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/openziti/zrok/agent/agentClient"
	"github.com/openziti/zrok/agent/agentGrpc"
	"github.com/openziti/zrok/endpoints"
	"github.com/openziti/zrok/endpoints/proxy"
	"github.com/openziti/zrok/endpoints/tcpTunnel"
	"github.com/openziti/zrok/endpoints/udpTunnel"
	"github.com/openziti/zrok/endpoints/vpn"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/environment/env_core"
	"github.com/openziti/zrok/rest_client_zrok"
	"github.com/openziti/zrok/rest_client_zrok/share"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/tui"
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
	headless        bool
	subordinate     bool
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
	cmd.Flags().BoolVar(&command.headless, "headless", false, "Disable TUI and run headless")
	cmd.Flags().BoolVar(&command.subordinate, "subordinate", false, "Enable subordinate mode")
	cmd.MarkFlagsMutuallyExclusive("headless", "subordinate")
	cmd.Flags().StringVarP(&command.bindAddress, "bind", "b", "127.0.0.1:9191", "The address to bind the private frontend")
	cmd.Flags().StringArrayVar(&command.responseHeaders, "response-header", []string{}, "Add a response header ('key:value')")
	cmd.Run = command.run
	return command
}

func (cmd *accessPrivateCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading environment", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	if cmd.subordinate {
		cmd.accessLocal(args, root)
	} else {
		agent, err := agentClient.IsAgentRunning(root)
		if err != nil {
			tui.Error("error checking if agent is running", err)
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
		if !panicInstead {
			tui.Error("unable to create zrok client", err)
		}
		panic(err)
	}

	auth := httptransport.APIKeyAuth("X-TOKEN", "header", root.Environment().Token)
	req := share.NewAccessParams()
	req.Body = &rest_model_zrok.AccessRequest{
		ShrToken: shrToken,
		EnvZID:   root.Environment().ZitiIdentity,
	}
	accessResp, err := zrok.Share.Access(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("unable to access", err)
		}
		panic(err)
	}

	if cmd.subordinate {
		data := make(map[string]interface{})
		data["frontend_token"] = accessResp.Payload.FrontendToken
		data["bind_address"] = cmd.bindAddress
		jsonData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
	} else {
		logrus.Infof("allocated frontend '%v'", accessResp.Payload.FrontendToken)
	}

	protocol := "http://"
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		protocol = "tcp://"
	case "udpTunnel":
		protocol = "udp://"
	}

	endpointUrl, err := url.Parse(protocol + cmd.bindAddress)
	if err != nil {
		if !panicInstead {
			tui.Error("invalid endpoint address", err)
		}
		panic(err)
	}

	requests := make(chan *endpoints.Request, 1024)
	switch accessResp.Payload.BackendMode {
	case "tcpTunnel":
		fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
			BindAddress:  cmd.bindAddress,
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private access", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting access", err)
				}
				panic(err)
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
			if !panicInstead {
				tui.Error("unable to create private frontend", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting frontend", err)
				}
				panic(err)
			}
		}()

	case "socks":
		fe, err := tcpTunnel.NewFrontend(&tcpTunnel.FrontendConfig{
			BindAddress:  cmd.bindAddress,
			IdentityName: root.EnvironmentIdentityName(),
			ShrToken:     args[0],
			RequestsChan: requests,
		})
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private access", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting access", err)
				}
				panic(err)
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
			if !panicInstead {
				tui.Error("unable to create private access", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("error starting access", err)
				}
				panic(err)
			}
		}()

	default:
		cfg := proxy.DefaultFrontendConfig(root.EnvironmentIdentityName())
		cfg.ShrToken = shrToken
		cfg.Address = cmd.bindAddress
		cfg.ResponseHeaders = cmd.responseHeaders
		cfg.RequestsChan = requests
		fe, err := proxy.NewFrontend(cfg)
		if err != nil {
			if !panicInstead {
				tui.Error("unable to create private frontend", err)
			}
			panic(err)
		}
		go func() {
			if err := fe.Run(); err != nil {
				if !panicInstead {
					tui.Error("unable to run frontend", err)
				}
			}
		}()
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		<-c
		cmd.destroy(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
		os.Exit(0)
	}()

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
		cmd.destroy(accessResp.Payload.FrontendToken, root.Environment().ZitiIdentity, shrToken, zrok, auth)
	}
}

func (cmd *accessPrivateCommand) destroy(frontendName, envZId, shrToken string, zrok *rest_client_zrok.Zrok, auth runtime.ClientAuthInfoWriter) {
	logrus.Infof("shutting down '%v'", shrToken)
	req := share.NewUnaccessParams()
	req.Body = &rest_model_zrok.UnaccessRequest{
		FrontendToken: frontendName,
		ShrToken:      shrToken,
		EnvZID:        envZId,
	}
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

	acc, err := client.AccessPrivate(context.Background(), &agentGrpc.AccessPrivateRequest{
		Token:           args[0],
		BindAddress:     cmd.bindAddress,
		ResponseHeaders: cmd.responseHeaders,
	})
	if err != nil {
		tui.Error("error creating access", err)
	}

	fmt.Println(acc)
}
