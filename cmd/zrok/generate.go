package main

import (
	"fmt"
	"strconv"

	"github.com/jaevor/go-nanoid"
	"github.com/openziti-test-kitchen/zrok/rest_client_zrok/invite"
	"github.com/openziti-test-kitchen/zrok/rest_model_zrok"
	"github.com/openziti-test-kitchen/zrok/zrokdir"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newGenerateCommand().cmd)
}

type generateCommand struct {
	cmd *cobra.Command
}

func newGenerateCommand() *generateCommand {
	cmd := &cobra.Command{
		Use:   "generate <optional-amount>",
		Short: "Generate invite tokens (default: 5)",
		Args:  cobra.RangeArgs(0, 1),
	}
	command := &generateCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *generateCommand) run(_ *cobra.Command, args []string) {
	var iterations int64 = 5
	if len(args) == 1 {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			showError("unable to parse amount", err)
		}
		iterations = i
	}
	var err error
	tokens := make([]string, iterations)
	for i := 0; i < int(iterations); i++ {
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
	req := invite.NewInviteGenerateParams()
	req.Body = &rest_model_zrok.InviteGenerateRequest{
		Tokens: tokens,
	}
	_, err = zrok.Invite.InviteGenerate(req)
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
