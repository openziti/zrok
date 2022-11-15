# v0.3.0 (WiP)

## Frontend Selection; Private Shares

The `zrok` model has been extended to include support for both a "public share" (exposing a backend through the globally-available `frontend` instances), and also a "private share" (exposing a backend service to a user who instantiates a private, local `frontend`).

### Underlying Schema Changes

* Added new `frontends` table
* Added new `availability_type` enumeration for use in the new `frontends` table
* Made the `account_id` column of the `environments` table `NULL`-able; a `NULL` value in the `account_id` column signifies an "ephemeral" environment

## Loop Test Shutdown Hook

The `zrok test loop` command now includes a shutdown hook to allow premature cancellation of a running test.

# v0.2.18

* First official release in the `v0.2.x` series. 
