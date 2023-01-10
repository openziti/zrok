package main

import (
	"fmt"

	"github.com/jaevor/go-nanoid"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/admin"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminGenerateCommand().cmd)
}

type adminGenerateCommand struct {
	cmd    *cobra.Command
	amount int
}

func newAdminGenerateCommand() *adminGenerateCommand {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate invite tokens (default: 5)",
		Args:  cobra.ExactArgs(0),
	}
	command := &adminGenerateCommand{cmd: cmd}
	cmd.Run = command.run

	cmd.Flags().IntVar(&command.amount, "amount", 5, "Amount of tokens to generate")

	return command
}

func (cmd *adminGenerateCommand) run(_ *cobra.Command, args []string) {
	var err error
	tokens := make([]string, cmd.amount)
	for i := 0; i < int(cmd.amount); i++ {
		tokens[i], err = createToken()
		if err != nil {
			showError("error creating token", err)
		}
	}
	zrok, err := zrokdir.ZrokClient(apiEndpoint)
	if err != nil {
		if !panicInstead {
			showError("error creating zrok api client", err)
		}
		panic(err)
	}
	req := admin.NewInviteTokenGenerateParams()
	req.Body = &rest_model_zrok.InviteTokenGenerateRequest{
		Tokens: tokens,
	}
	_, err = zrok.Admin.InviteTokenGenerate(req, mustGetAdminAuth())
	if err != nil {
		if !panicInstead {
			showError("error creating invite tokens", err)
		}
		panic(err)
	}
	fmt.Printf("generated %d tokens\n", len(tokens))
}

func createToken() (string, error) {
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)
	if err != nil {
		return "", err
	}
	return gen(), nil
}
