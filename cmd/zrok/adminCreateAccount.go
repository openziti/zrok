package main

import (
	"fmt"
	"github.com/openziti/zrok/controller"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateAccount().cmd)
}

type adminCreateAccount struct {
	cmd *cobra.Command
}

func newAdminCreateAccount() *adminCreateAccount {
	cmd := &cobra.Command{
		Use:   "account <configPath}> <email> <password>",
		Short: "Pre-populate an account in the database; returns an enable token for the account",
		Args:  cobra.ExactArgs(3),
	}
	command := &adminCreateAccount{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateAccount) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	str, err := store.Open(cfg.Store)
	if err != nil {
		panic(err)
	}
	token, err := controller.CreateToken()
	if err != nil {
		panic(err)
	}
	hpwd, err := controller.HashPassword(args[2])
	if err != nil {
		panic(err)
	}
	trx, err := str.Begin()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := trx.Commit(); err != nil {
			panic(err)
		}
	}()
	a := &store.Account{
		Email:    args[1],
		Salt:     hpwd.Salt,
		Password: hpwd.Password,
		Token:    token,
	}
	if _, err := str.CreateAccount(a, trx); err != nil {
		panic(err)
	}
	fmt.Println(token)
}
