---
sidebar_position: 21
sidebar_label: Organizations
---

# Organizations

zrok (starting with `v0.4.45`) includes support for "organizations". Organizations are groups of related accounts that are typically centrally managed in some capacity. A zrok account can be a member of multiple organizations. Organization membership can also include an "admin" permission. As of `v0.4.45` organization admins are able to retrieve an "overview" (`zrok overview`) from any other account in the organization, allowing the admin to see the details of the environments, shares, and accesses created within that account.

Future zrok releases will include additional organization features, including `--closed` permission sharing functions.

## Configuring an Organization

The API endpoints used to manage organizations and their members require a site-level `ZROK_ADMIN_TOKEN` to access. See the [self-hosting guide](linux/index.mdx#configure-the-controller) for details on configuring admin tokens.

### Create an Organization

The `zrok admin create organization` command is used to create organizations:

```
$ zrok admin create organization --help
Create a new organization

Usage:
  zrok admin create organization [flags]

Aliases:
  organization, org

Flags:
  -d, --description string   Organization description
  -h, --help                 help for organization

Global Flags:
  -p, --panic     Panic instead of showing pretty errors
  -v, --verbose   Enable verbose logging
```

Use the `-d` flag to add a description that shows up in end-user membership listings.

We'll create an example organization:

```
$ zrok admin create organization -d "documentation"
[   0.006]    INFO main.(*adminCreateOrganizationCommand).run: created new organization with token 'gK1XRvthq7ci'
```

### List Organizations

We use the `zrok admin list organizations` command to list our organizations:

```
$ zrok admin list organizations

 ORGANIZATION TOKEN  DESCRIPTION   
 gK1XRvthq7ci        documentation 
```

### Add a Member to an Organization

We use the `zrok admin create org-member` command to add members to organizations:

```
$ zrok admin create org-member 
Error: accepts 2 arg(s), received 0
Usage:
  zrok admin create org-member <organizationToken> <accountEmail> [flags]

Aliases:
  org-member, member

Flags:
      --admin   Make the new account an admin of the organization
  -h, --help    help for org-member

Global Flags:
  -p, --panic     Panic instead of showing pretty errors
  -v, --verbose   Enable verbose logging
```

Like this:

```
$ zrok admin create org-member gK1XRvthq7ci michael.quigley@netfoundry.io
[   0.006]    INFO main.(*adminCreateOrgMemberCommand).run: added 'michael.quigley@netfoundry.io' to organization 'gK1XRvthq7ci
```

The `--admin` flag can be added to the `zrok admin create org-member` command to mark the member as an administrator of the organization.

### List Members of an Organization

```
$ zrok admin list org-members gK1XRvthq7ci

 ACCOUNT EMAIL                  ADMIN? 
 michael.quigley@netfoundry.io  false 
```

### Removing Organizations and Members

The `zrok admin delete org-member` and `zrok admin delete organization` commands are available to clean up organizations and their membership lists.

## End-user Organization Administrator Commands

When a zrok account is added to an organization as an administrator it allows them to use the `zrok organization admin` commands, which include:

```
$ zrok organization admin
Organization admin commands

Usage:
  zrok organization admin [command]

Available Commands:
  list        List the members of an organization
  overview    Retrieve account overview for organization member account

Flags:
  -h, --help   help for admin

Global Flags:
  -p, --panic     Panic instead of showing pretty errors
  -v, --verbose   Enable verbose logging

Use "zrok organization admin [command] --help" for more information about a command.
```

The `zrok organization admin list` command is used to list the members of an organization.

The `zrok organization admin overview` command is used to retrieve an overview of an organization member account. This is functionally equivalent to what the `zrok overview` command does, but it allows an organization admin to retrieve the overview for another zrok account.

## End-user Organization Commands

All zrok accounts can use the `zrok organization memberships` command to list the organizations they're a member of:

```
$ zrok organization memberships

 ORGANIZATION TOKEN  DESCRIPTION    ADMIN? 
 gK1XRvthq7ci        documentation  false  

```



