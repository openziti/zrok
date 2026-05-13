package main

import (
	"strconv"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminDeleteCmd.AddCommand(newAdminDeleteAppliedLimitClassesCommand().cmd)
}

type adminDeleteAppliedLimitClassesCommand struct {
	cmd *cobra.Command
}

func newAdminDeleteAppliedLimitClassesCommand() *adminDeleteAppliedLimitClassesCommand {
	cmd := &cobra.Command{
		Use:     "applied-limit-classes <email> <limitClassId> [<limitClassId>...]",
		Aliases: []string{"alcs"},
		Short:   "Remove one or more applied limit classes from the specified account",
		Args:    cobra.MinimumNArgs(2),
	}
	command := &adminDeleteAppliedLimitClassesCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminDeleteAppliedLimitClassesCommand) run(_ *cobra.Command, args []string) {
	env, err := environment.LoadRoot()
	if err != nil {
		panic(err)
	}

	zrok, err := env.Client()
	if err != nil {
		panic(err)
	}

	var limitClassIds []int64
	for _, arg := range args[1:] {
		lcId, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			panic(err)
		}
		limitClassIds = append(limitClassIds, lcId)
	}

	req := admin.NewRemoveAppliedLimitClassesParams()
	req.Body.Email = args[0]
	req.Body.LimitClassIds = limitClassIds

	_, err = zrok.Admin.RemoveAppliedLimitClasses(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("removed %d applied limit class(es) from '%v'", len(limitClassIds), args[0])
}
