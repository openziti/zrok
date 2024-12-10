package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

func init() {
	organizationCmd.AddCommand(newOrgAccountOverviewCommand().cmd)
}

type orgAccountOverviewCommand struct {
	cmd *cobra.Command
}

func newOrgAccountOverviewCommand() *orgAccountOverviewCommand {
	cmd := &cobra.Command{
		Use:     "account-overview <organizationToken> <accountEmail>",
		Aliases: []string{"overview"},
		Short:   "Retrieve account overview for organization account (requires admin)",
		Args:    cobra.ExactArgs(2),
	}
	command := &orgAccountOverviewCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *orgAccountOverviewCommand) run(_ *cobra.Command, args []string) {
	root, err := environment.LoadRoot()
	if err != nil {
		if !panicInstead {
			tui.Error("error loading zrokdir", err)
		}
		panic(err)
	}

	if !root.IsEnabled() {
		tui.Error("unable to load environment; did you 'zrok enable'?", nil)
	}

	client := &http.Client{}
	apiEndpoint, _ := root.ApiEndpoint()
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v1/overview/%v/%v", apiEndpoint, args[0], args[1]), nil)
	if err != nil {
		if !panicInstead {
			tui.Error("error creating request", err)
		}
		panic(err)
	}
	req.Header.Add("X-TOKEN", root.Environment().Token)
	resp, err := client.Do(req)
	if err != nil {
		if !panicInstead {
			tui.Error("error sending request", err)
		}
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		if !panicInstead {
			tui.Error("received error response", errors.New(resp.Status))
		}
		panic(errors.New(resp.Status))
	}

	json, err := io.ReadAll(resp.Body)
	if err != nil {
		if !panicInstead {
			tui.Error("error reading json", err)
		}
		panic(err)
	}
	_ = resp.Body.Close()

	fmt.Println(string(json))
}
