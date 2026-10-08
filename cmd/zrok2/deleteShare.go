package main

import (
	"sort"
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/util"
	"github.com/spf13/cobra"
)

func init() {
	deleteCmd.AddCommand(newDeleteShareCommand().cmd)
}

type deleteShareCommand struct {
	cmd    *cobra.Command
	envZId string
}

func newDeleteShareCommand() *deleteShareCommand {
	cmd := &cobra.Command{
		Use:   "share <shareToken|name>",
		Short: "Delete a share",
		Long: "Delete a share, given its share token or a name it holds. A name is given as <name>, " +
			"<namespaceToken>:<name> or <name>.<namespaceName>; the qualified forms pick one namespace when the " +
			"name is held in several.\n\n" + exitCodesHelp,
		Args: cobra.ExactArgs(1),
	}
	command := &deleteShareCommand{cmd: cmd}
	cmd.Flags().StringVar(&command.envZId, "envzid", "", "Override environment ziti identifier")
	cmd.Run = command.run
	return command
}

func (cmd *deleteShareCommand) run(_ *cobra.Command, args []string) {
	env, auth := mustGetEnvironmentAuth()
	zrok, err := env.Client()
	if err != nil {
		exitWithFailure("unable to create zrok client", err)
	}

	sharesResp, err := zrok.Metadata.ListShares(metadata.NewListSharesParams(), auth)
	if err != nil {
		exitWithFailure("unable to list shares", err)
	}
	var names []*rest_model_zrok.Name
	if !isShareToken(args[0], sharesResp.Payload.Shares) {
		namesResp, err := zrok.Share.ListAllNames(share.NewListAllNamesParams(), auth)
		if err != nil {
			exitWithFailure("unable to list names", err)
		}
		names = namesResp.Payload
	}
	shrToken, err := resolveShareTokenOrName(args[0], sharesResp.Payload.Shares, names)
	if err != nil {
		exitWithFailure("unable to delete share", err)
	}

	req := share.NewUnshareParams()
	req.Body.EnvZID = resolveUnshareEnvZId(args[0], shrToken, env.Environment().ZitiIdentity, cmd.envZId, sharesResp.Payload.Shares)
	req.Body.ShareToken = shrToken

	_, err = zrok.Share.Unshare(req, auth)
	if err != nil {
		exitWithFailure("unable to delete share", err)
	}

	dl.Infof("deleted share '%v' from environment '%v'", req.Body.ShareToken, req.Body.EnvZID)
}

// resolveUnshareEnvZId picks the environment to unshare from: --envzid when given; for a share resolved
// from a name, the environment that share lives in, which need not be the current one; otherwise the
// current environment, as for a share token.
func resolveUnshareEnvZId(arg, shrToken, currentEnvZId, overrideEnvZId string, shares []*rest_model_zrok.ShareSummary) string {
	if overrideEnvZId != "" {
		return overrideEnvZId
	}
	if arg != shrToken {
		for _, shr := range shares {
			if shr != nil && shr.ShareToken == shrToken && shr.EnvZID != "" {
				return shr.EnvZID
			}
		}
	}
	return currentEnvZId
}

func isShareToken(arg string, shares []*rest_model_zrok.ShareSummary) bool {
	for _, shr := range shares {
		if shr != nil && shr.ShareToken == arg {
			return true
		}
	}
	return false
}

// resolveShareTokenOrName returns arg when it is one of the account's share tokens, and otherwise the token
// of the live share holding the name arg. the names listing reports only live shares' tokens.
func resolveShareTokenOrName(arg string, shares []*rest_model_zrok.ShareSummary, names []*rest_model_zrok.Name) (string, error) {
	if isShareToken(arg, shares) {
		return arg, nil
	}

	var matched []*rest_model_zrok.Name
	for _, name := range names {
		if name != nil && nameMatches(arg, name) {
			matched = append(matched, name)
		}
	}
	if len(matched) == 0 {
		return "", newPermanentError("no share token or name '%v' found for this account", arg)
	}

	held := make(map[string][]string)
	for _, name := range matched {
		if name.ShareToken != "" {
			held[name.ShareToken] = append(held[name.ShareToken], name.NamespaceToken+":"+name.Name)
		}
	}
	switch len(held) {
	case 0:
		return "", newPermanentError("name '%v' is not held by a live share", arg)
	case 1:
		for shrToken := range held {
			return shrToken, nil
		}
	}

	var holders []string
	for shrToken, qualified := range held {
		holders = append(holders, "'"+strings.Join(qualified, "', '")+"' (share '"+shrToken+"')")
	}
	sort.Strings(holders)
	return "", newPermanentError("name '%v' is held by more than one share: %v; qualify it as <namespaceToken>:<name>", arg, strings.Join(holders, ", "))
}

// nameMatches reports whether arg names name, as <name>, <namespaceToken>:<name> or <name>.<namespaceName>.
func nameMatches(arg string, name *rest_model_zrok.Name) bool {
	return arg == name.Name ||
		arg == name.NamespaceToken+":"+name.Name ||
		(name.NamespaceName != "" && arg == util.NameInNamespace(name.Name, name.NamespaceName))
}
