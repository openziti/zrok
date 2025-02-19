package main

import (
	"fmt"
	"github.com/jaevor/go-nanoid"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/rest_client_zrok/admin"
	"github.com/sirupsen/logrus"
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
		Short: "Generate invite tokens",
		Args:  cobra.ExactArgs(0),
	}
	command := &adminGenerateCommand{cmd: cmd}
	cmd.Run = command.run
	cmd.Flags().IntVarP(&command.amount, "count", "n", 5, "Number of tokens to generate")
	return command
}

func (cmd *adminGenerateCommand) run(_ *cobra.Command, args []string) {
	var err error
	tokens := make([]string, cmd.amount)
	for i := 0; i < int(cmd.amount); i++ {
		tokens[i], err = createToken()
		if err != nil {
			logrus.Error("error creating token", err)
		}
	}

	env, err := environment.LoadRoot()
	if err != nil {
		logrus.Error("error loading environment", err)
	}

	zrok, err := env.Client()
	if err != nil {
		if !panicInstead {
			logrus.Error("error creating zrok api client", err)
		}
		panic(err)
	}
	req := admin.NewInviteTokenGenerateParams()
	req.Body.InviteTokens = tokens

	_, err = zrok.Admin.InviteTokenGenerate(req, mustGetAdminAuth())
	if err != nil {
		if !panicInstead {
			logrus.Error("error creating invite tokens", err)
		}
		panic(err)
	}

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func createToken() (string, error) {
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)
	if err != nil {
		return "", err
	}
	return gen(), nil
}
