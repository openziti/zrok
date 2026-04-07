---
sidebar_position: 21
sidebar_label: Organizations
---

# Organizations

zrok (starting with `v0.4.45`) includes support for organizations—groups of related accounts that are typically
centrally managed. A zrok account can be a member of multiple organizations, and membership can include an admin
permission. Organization admins can retrieve an overview (`zrok overview`) for any other account in the organization,
giving them visibility into the environments, shares, and accesses created within that account.

## Configure an organization

Managing organizations requires a site-level `ZROK2_ADMIN_TOKEN`. See the
[self-hosting guide](deployment/linux.mdx#step-2-configure-the-controller) for details on configuring admin tokens.

### Create an organization

Use `zrok2 admin create organization` to create an organization:

```bash
zrok2 admin create organization --help
```

```buttonless title="Output"
Create a new organization

Usage:
  zrok2 admin create organization [flags]

Aliases:
  organization, org

Flags:
  -d, --description string   Organization description
  -h, --help                 help for organization

Global Flags:
  -p, --panic     Panic instead of showing pretty errors
  -v, --verbose   Enable verbose logging
```

Use the `-d` flag to add a description that shows up in end-user membership listings. For example:

```bash
zrok2 admin create organization -d "documentation"
```

```buttonless title="Output"
[   0.006]    INFO main.(*adminCreateOrganizationCommand).run: created new organization with token 'gK1XRvthq7ci'
```

### List organizations

Use `zrok2 admin list organizations` to list organizations:

```bash
zrok2 admin list organizations
```

```buttonless title="Output"
 ORGANIZATION TOKEN  DESCRIPTION   
 gK1XRvthq7ci        documentation 
```

### Add a member to an organization

Use `zrok2 admin create org-member` to add a member to an organization:

```bash
zrok2 admin create org-member --help
```

```buttonless title="Output"
Usage:
  zrok2 admin create org-member <organizationToken> <accountEmail> [flags]

Aliases:
  org-member, member

Flags:
      --admin   Make the new account an admin of the organization
  -h, --help    help for org-member

Global Flags:
  -p, --panic     Panic instead of showing pretty errors
  -v, --verbose   Enable verbose logging
```

Add the `--admin` flag to mark the member as an organization administrator. For example:

```bash
zrok2 admin create org-member gK1XRvthq7ci michael.quigley@netfoundry.io
```

```buttonless title="Output"
[   0.006]    INFO main.(*adminCreateOrgMemberCommand).run: added 'michael.quigley@netfoundry.io' to organization 'gK1XRvthq7ci'
```

### List members of an organization

Use `zrok2 admin list org-members <organizationToken>` to list the members of an organization:

```bash
zrok2 admin list org-members gK1XRvthq7ci
```

```buttonless title="Output"
 ACCOUNT EMAIL                  ADMIN? 
 michael.quigley@netfoundry.io  false 
```

### Remove organizations and members

Use `zrok2 admin delete org-member` and `zrok2 admin delete organization` to remove members and organizations.

## End-user organization administrator commands

Organization admins can use the `zrok2 organization admin` commands:

```bash
zrok2 organization admin --help
```

```buttonless title="Output"
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

Use `zrok2 organization admin list` to list the members of an organization.

Use `zrok2 organization admin overview` to retrieve an overview of a member account. This works like `zrok2 overview`
but lets an organization admin retrieve the overview for any member account.

## End-user organization commands

Use `zrok2 organization memberships` to list the organizations your account belongs to:

```bash
zrok2 organization memberships
```

```buttonless title="Output"
 ORGANIZATION TOKEN  DESCRIPTION    ADMIN? 
 gK1XRvthq7ci        documentation  false  
```
