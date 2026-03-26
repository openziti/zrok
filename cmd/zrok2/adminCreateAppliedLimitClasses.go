package main

import (
	"strconv"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/admin"
	"github.com/spf13/cobra"
)

func init() {
	adminCreateCmd.AddCommand(newAdminCreateAppliedLimitClassesCommand().cmd)
}

type adminCreateAppliedLimitClassesCommand struct {
	cmd *cobra.Command
}

func newAdminCreateAppliedLimitClassesCommand() *adminCreateAppliedLimitClassesCommand {
	cmd := &cobra.Command{
		Use:     "applied-limit-classes <email> <limitClassId> [<limitClassId>...]",
		Aliases: []string{"alcs"},
		Short:   "Apply one or more limit classes to the specified account",
		Args:    cobra.MinimumNArgs(2),
	}
	command := &adminCreateAppliedLimitClassesCommand{cmd: cmd}
	cmd.Run = command.run
	return command
}

func (cmd *adminCreateAppliedLimitClassesCommand) run(_ *cobra.Command, args []string) {
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

	req := admin.NewApplyLimitClassesParams()
	req.Body.Email = args[0]
	req.Body.LimitClassIds = limitClassIds

	_, err = zrok.Admin.ApplyLimitClasses(req, mustGetAdminAuth())
	if err != nil {
		panic(err)
	}

	dl.Infof("applied %d limit class(es) to '%v'", len(limitClassIds), args[0])
}
