package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/michaelquigley/df/dl"
	restEnvironment "github.com/openziti/zrok/v2/rest_client_zrok/environment"
	"github.com/openziti/zrok/v2/tui"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(newDeleteEnvironmentCommand().cmd)
}

type deleteEnvironmentCommand struct {
	cmd          *cobra.Command
	accountToken string
	force        bool
}

func newDeleteEnvironmentCommand() *deleteEnvironmentCommand {
	cmd := &cobra.Command{
		Use:   "environment <envZId>",
		Short: "Delete a zrok environment by its envZId",
		Args:  cobra.ExactArgs(1),
		Long: `Delete a zrok environment other than your current environment.

If you want to delete your current environment, use 'zrok2 disable' instead,
which will also clean up local configuration files.

This command requires a local environment to determine the API endpoint.
By default, it uses the local environment's account token for authentication.
Use --account-token to authenticate as a different account.

Examples:
  zrok2 delete environment abc123def456
  zrok2 delete environment abc123def456 --force
  zrok2 delete environment abc123def456 --account-token your-token-here
`,
	}
	command := &deleteEnvironmentCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.accountToken, "account-token", "", "account token for authentication (use when no enabled environment exists)")
	cmd.Flags().BoolVarP(&command.force, "force", "f", false, "skip confirmation prompt")
	cmd.Run = command.run
	return command
}

func (cmd *deleteEnvironmentCommand) run(_ *cobra.Command, args []string) {
	envZIdToDelete := args[0]

	// get environment and auth (either from local env or provided token)
	env, auth, err := getEnvironmentAuthOptional(cmd.accountToken)
	if err != nil {
		if !panicInstead {
			tui.Error("authentication error", err)
		}
		panic(err)
	}

	// if we have an enabled environment, check if user is trying to delete it
	if env.IsEnabled() && env.Environment().ZitiIdentity == envZIdToDelete {
		if !panicInstead {
			tui.Error("cannot delete current environment", fmt.Errorf("you are trying to delete your current environment; use 'zrok2 disable' instead to properly clean up local files"))
		}
		panic("cannot delete current environment; use 'zrok2 disable' instead")
	}

	// confirmation prompt (unless --force is used)
	if !cmd.force {
		if !cmd.confirmDeletion(envZIdToDelete) {
			fmt.Println("deletion cancelled")
			return
		}
	}

	// get API client (env is always available from getEnvironmentAuthOptional)
	zrok, err := env.Client()
	if err != nil {
		if !panicInstead {
			tui.Error("could not create zrok client", err)
		}
		panic(err)
	}

	// call the disable endpoint with the specified envZId
	req := restEnvironment.NewDisableParams()
	req.Body.Identity = envZIdToDelete

	_, err = zrok.Environment.Disable(req, auth)
	if err != nil {
		if !panicInstead {
			tui.Error("failed to delete environment", err)
		}
		panic(err)
	}

	dl.Infof("deleted environment '%v'", envZIdToDelete)
}

// confirmDeletion prompts the user for confirmation and returns true if they confirm
func (cmd *deleteEnvironmentCommand) confirmDeletion(envZId string) bool {
	fmt.Printf("are you sure you want to delete environment '%s'? [y/N]: ", envZId)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
