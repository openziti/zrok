# Account Request Process

## In v0.1

The `v0.1` versions of `zrok` had an open-access `zrok create account` that allows any user to create an account. Useful for closed development environments only.

## In v0.2

* The `zrok create account` command now only takes an email address. 
* The email address is submitted to an open-ended API endpoint, which then records an "account request", allocating a request token.
* An email is sent to the address offering a link with the request token, allowing the user to create the account.
* The account request is marked complete.

### Invitations for Others

This open `zrok create account` command will allow any user to send a `zrok` invitation to any user with a valid email address.

### Garbage Collection

An background garbage collector in the controller scans the account requests, looking for unused requests, which are removed after a configurable amount of time.