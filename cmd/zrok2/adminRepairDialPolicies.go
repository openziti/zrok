package main

import (
	"github.com/michaelquigley/df/dd"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller"
	"github.com/openziti/zrok/v2/controller/config"
	"github.com/spf13/cobra"
)

func init() {
	adminCmd.AddCommand(newAdminRepairDialPoliciesCommand().cmd)
}

type adminRepairDialPoliciesCommand struct {
	cmd   *cobra.Command
	apply bool
}

func newAdminRepairDialPoliciesCommand() *adminRepairDialPoliciesCommand {
	cmd := &cobra.Command{
		Use:   "repair-dial-policies <configPath>",
		Short: "Restore missing dial service policies of unlimited shares (dry run by default)",
		Long: "Report, and with --apply create, the dial service policies that live shares should have and do not:\n" +
			"a public share's frontend policy, derived from its name mappings, and the policy of each private access.\n\n" +
			"Without --apply this is a dry run: nothing is created and the report lists what is missing.\n" +
			"Accounts the bandwidth limit journal holds limited are skipped, as are shares of a backend mode a\n" +
			"scoped limit class holds limited. Nothing is ever deleted, and a second run finds nothing missing.",
		Args: cobra.ExactArgs(1),
	}
	command := &adminRepairDialPoliciesCommand{cmd: cmd}
	cmd.Flags().BoolVar(&command.apply, "apply", false, "Create the missing dial policies (default is a dry run)")
	cmd.Run = command.run
	return command
}

func (c *adminRepairDialPoliciesCommand) run(_ *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(args[0])
	if err != nil {
		panic(err)
	}
	dl.Info(dd.MustInspect(cfg))
	if err := controller.RepairDialPolicies(cfg, c.apply); err != nil {
		panic(err)
	}
}
