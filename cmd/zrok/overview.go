package main

import (
	"fmt"
	"github.com/openziti/zrok/environment"
	"github.com/openziti/zrok/tui"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

func init() {
	rootCmd.AddCommand(newOverviewCommand().cmd)
}

type overviewCommand struct {
	cmd *cobra.Command
}

func newOverviewCommand() *overviewCommand {
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Retrieve all of the zrok account details (environments, shares) as JSON",
		Args:  cobra.ExactArgs(0),
	}
	command := &overviewCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *overviewCommand) run(_ *cobra.Command, _ []string) {
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/api/v1/overview", apiEndpoint), nil)
	if err != nil {
		if !panicInstead {
			tui.Error("error accessing overview", err)
		}
		panic(err)
	}
	req.Header.Add("X-TOKEN", root.Environment().Token)
	resp, err := client.Do(req)
	if err != nil {
		if !panicInstead {
			tui.Error("error requesting overview", err)
		}
		panic(err)
	}

	json, err := io.ReadAll(resp.Body)
	if err != nil {
		if !panicInstead {
			tui.Error("error reading body", err)
		}
		panic(err)
	}
	_ = resp.Body.Close()

	fmt.Println(string(json))
}
