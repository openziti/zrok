package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/secretsGrpc"
	"github.com/openziti/zrok/environment"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func init() {
	adminListCmd.AddCommand(newAdminListSecretsCommand().cmd)
}

type adminListSecretsCommand struct {
	cmd *cobra.Command
}

func newAdminListSecretsCommand() *adminListSecretsCommand {
	cmd := &cobra.Command{
		Use:   "secrets <secretsAccessIdentity> <serviceName> <shareToken>",
		Short: "Retrieve secrets from the secrets store",
		Args:  cobra.ExactArgs(3),
	}
	command := &adminListSecretsCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminListSecretsCommand) run(_ *cobra.Command, args []string) {
	secretsAccessIdentityName := args[0]
	serviceName := args[1]
	shareToken := args[2]

	client, conn, err := cmd.newSecretsClient(secretsAccessIdentityName, serviceName)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	resp, err := client.FetchSecrets(context.Background(), &secretsGrpc.SecretsRequest{
		ShareToken: shareToken,
	})
	if err != nil {
		panic(err)
	}

	for _, secret := range resp.Secrets {
		fmt.Printf("%v: %v\n", secret.Key, secret.Value)
	}
}

func (cmd *adminListSecretsCommand) newSecretsClient(secretsAccessIdentityName, serviceName string) (client secretsGrpc.SecretsClient, conn *grpc.ClientConn, err error) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}
	zif, err := env.ZitiIdentityNamed(secretsAccessIdentityName)
	if err != nil {
		return nil, nil, err
	}
	opts := []grpc.DialOption{
		grpc.WithContextDialer(func(_ context.Context, addr string) (net.Conn, error) {
			zcfg, err := ziti.NewConfigFromFile(zif)
			if err != nil {
				return nil, err
			}
			zctx, err := ziti.NewContext(zcfg)
			if err != nil {
				return nil, err
			}
			conn, err := zctx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
			if err != nil {
				return nil, err
			}
			return conn, nil
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	resolver.SetDefaultScheme("passthrough")
	conn, err = grpc.NewClient(serviceName, opts...)
	if err != nil {
		return nil, nil, err
	}
	return secretsGrpc.NewSecretsClient(conn), conn, nil
}
