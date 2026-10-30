# Agent smoke

The `agent smoke` workflow (`.github/workflows/agent-smoke.yml`) runs the agent's named public share through an interrupt and a restart on native GitHub-hosted runners, against a live zrok v2 service with a dedicated test account. It builds `zrok2` from the triggering commit on each runner and drives it as a user would. It supplies the native platform evidence recorded in [Agent share lifecycle](agent-share-lifecycle.md). It is the only CI workflow that uses a live service account. It runs only on manual dispatch. No other workflow, required check, release or `Makefile` target depends on it.

One platform matrix runs the same build, fixture, live-scenario and cleanup sequence on each row, each on a fresh runner:

| Row | Runner | Architecture | Graceful stop |
| --- | --- | --- | --- |
| macOS arm64 | `macos-15` | arm64 | SIGINT to the agent's process group |
| Linux amd64 | `ubuntu-24.04` | amd64 | SIGINT to the agent's process group |
| Windows amd64 | `windows-2025` | amd64 | `CTRL_BREAK_EVENT` to the agent's console process group |

The scope of the graceful-stop event differs by platform. On POSIX the agent starts its workers in their own process groups, so SIGINT reaches only the agent, which stops them. On Windows the agent starts workers in its own console process group, so Ctrl-Break reaches the agent and its workers together, as a Ctrl-C typed into the agent's console would. Each worker stops on its own interrupt handler. Separate fixtures cover the agent-first worker stop that closes a worker's control input. The Windows row exercises Ctrl-Break, not a literal keyboard Ctrl-C: a targeted `CTRL_C_EVENT` cannot reach a new process group. Go delivers both events to the same interrupt handler.

## What a run proves

The live scenario passes only when every observation holds and final cleanup succeeds:

1. A fresh environment is enabled with the built CLI's headless enable path, and a new unique name is reserved.
2. `zrok2 agent start` runs in its own process group (on Windows, its own console process group inside a job object) and reports the exact build identity the job stamped.
3. `zrok2 share public --force-agent --open -n <namespace>:<name>` against a loopback server produces exactly one active share in structured agent status, the name is bound to its token, and the public endpoint returns the server's per-run random body. Bare endpoint hostnames are fetched over HTTPS; explicit HTTP or HTTPS schemes are preserved.
4. The row's graceful-stop event ends the agent with exit 0 within 130 seconds. On macOS and Linux that event is SIGINT to the agent's process group, as a terminal Ctrl-C delivers it; on Windows it is `CTRL_BREAK_EVENT`. A forced kill is a failure.
5. A successful authenticated account overview still lists the environment, no longer mentions the share token anywhere, and the name listing shows the name reserved and unbound.
6. Restarting the agent with the same environment replays the saved request without another share command: exactly one active share, a new token, the old token still absent, the name bound to the replacement and the random body served through it.
7. A second interrupt repeats the exit, absence and name checks. Cleanup then removes the name and the environment and proves their absence.

Before the live scenario each row runs the native process and worker fixtures (`./agent/proctree`, `./cmd/zrok2/subordinate`), and the agent package's lifecycle tests and the harness's offline tests under the race detector. The offline tests include the platform's own process fixtures. A helper that installs Go's interrupt handler must acknowledge the graceful-stop event and exit 0, through the same delivery the live scenario uses. Forced termination must also reach a worker started by the agent's own launcher, and on Windows a descendant orphaned by its parent. The POSIX signal fixtures in `./agent` skip Windows with their stated reasons.

A successful build, passing native fixtures or a run rescued by fallback cleanup does not close a platform's gap; only a row whose scenario steps and run-mode cleanup all pass does. Each row's result stands alone: a passing macOS row is no evidence for Linux or Windows, and a row cancelled while queued is not a result.

## Setup

The operator does these once, in repository settings. No step needs an admin token.

- A dedicated v2 account on the target service, used for nothing else, permitted to enable an environment and reserve a name in the target namespace. The job never provisions or modifies the account itself.
- A GitHub environment named `zrok-agent-smoke`, with deployment branches restricted to trusted branches, holding the secret `ZROK2_CI_ACCOUNT_TOKEN` (the dedicated account's token). An environment secret reaches only jobs that reference that environment, and within the job only the live scenario and fallback cleanup steps receive it.
- Optionally, environment variables `ZROK2_SMOKE_API_ENDPOINT` and `ZROK2_SMOKE_NAMESPACE` to target a dedicated service. The defaults are `https://api-v2.zrok.io` and `public`.

Windows runners have no configured C compiler for cgo and the race detector. That row installs MSYS2's UCRT64 gcc with the runner's preinstalled MSYS2 (`C:\msys64`), puts its `bin` directory on `PATH` for the remaining steps and sets `CC=gcc`. Each row's job summary names the compiler that Go used.

A missing secret fails the live step with an error naming it; no row is green without the live scenario.

## Dispatch

The temporary `fixes-209` push trigger was removed after the operator reported successful runs on all three platforms. Pushing a commit no longer starts this workflow.

GitHub's manual Run workflow button requires the workflow on the default branch. Once it reaches `main`, an authorized operator selects the branch to test: Actions, `agent smoke`, Run workflow, or `gh workflow run agent-smoke.yml --ref <branch>`. The branch must be one the `zrok-agent-smoke` environment admits. The workflow takes no inputs: the checkout is the triggering ref, and the endpoint and namespace come only from the environment's variables.

Matrix rows run one at a time, and a failed row does not cancel the rest. Runs share one concurrency group, and a run in progress is never cancelled by another; a second run waits for the first. GitHub keeps at most one waiting run per group, so a third run cancels the waiting one, and a cancelled run supplies no result for any row. The group keeps its original name, `macos-agent-smoke-zrok2-ci-account`, so a run of the earlier macOS-only workflow from an older branch also waits. Each row's job is bounded at 45 minutes, the live step at 20 and fallback cleanup at 5.

The build version is the latest `v2.*` tag reachable from the checkout, and `build.Hash` is the full commit hash. Both are written to each row's job summary, with the runner's OS release and architecture, the Go toolchain and the C compiler. The CLI and the harness are built with identical stamps, so the controller's version check applies to both. A version refusal fails the run as a failed prerequisite and keeps the controller's message.

## The harness

`tests/agent-smoke` is a Go command whose sources all carry the `agent_smoke` build tag. Process control is split into `process_unix.go` and `process_windows.go`; everything else is shared. Ordinary `go build ./...`, `go test ./...`, `go vet ./...`, `go install ./...` and `make test` never select it. Build it with `go build -tags agent_smoke ./tests/agent-smoke`.

```
agent-smoke run --binary <zrok2> --state-dir <dir> [--api-endpoint <url>] [--namespace <token>]
agent-smoke cleanup --state-dir <dir>
```

Both modes read the account token from `ZROK2_CI_ACCOUNT_TOKEN`. Only `run` creates resources. It refuses before any side effect when `~/.zrok2` exists, when the state directory already holds a manifest, or when the harness was built without `build.Version` and `build.Hash`. The CLI it drives always uses `~/.zrok2`, so a live run needs a fresh home: the hosted runner's, or a disposable OS user's when run locally. The harness never adopts, modifies or cleans an existing environment. The offline tests run anywhere: `go test -tags agent_smoke ./tests/agent-smoke`. They use a fake controller, a fake public frontend and the test binary re-executed as a fake `zrok2`, and never contact a zrok service.

Before any side effect, `run` chooses the name `agentsmoke-<run id>-<attempt>-<random>` and the environment description `zrok agent smoke <run id>-<attempt>-<random>`, from `GITHUB_RUN_ID` and `GITHUB_RUN_ATTEMPT` or `local-0` outside Actions. It records them in `<state-dir>/manifest.json` (mode 0600 on POSIX; Windows keeps it in the runner-owned state directory). It then adds each identifier as it is acquired: the environment's Ziti identity, every share token observed and each agent's pid and identity. A process record's `platform` says which identity it holds: `posix` (the start time and command line `ps` reports; also assumed for records without the field) or `windows` (the creation time and executable path read from a process handle). The manifest never holds the account token or identity material. A name already listed for the account fails the run, and that name is marked foreign so cleanup never touches it.

The scenario has an overall 15-minute bound. Agent readiness and share startup or replay each get at most 7 minutes, an agent stop 130 seconds and public HTTP convergence 120 seconds. Each controller or agent request gets 30 seconds, and each short CLI invocation 90 seconds. All of these are capped by the scenario's remaining time. Readiness is polled, never inferred from a sleep. The harness's own assertions read the generated REST client and the agent's structured gRPC status, independently of the agent's cleanup code.

Child output is held in memory (the first 64 KiB and last 192 KiB per process) and written to `<state-dir>/report/logs/` only after the account token has been redacted. The report directory also holds `run-summary.json`/`.md` and `cleanup-summary.json`/`.md`. It is each row's only uploaded artifact, named `agent-smoke-<row>-<run id>-<attempt>` (`macos-arm64`, `linux-amd64`, `windows-amd64`) and kept for seven days, and its markdown summary, which names the harness's OS and architecture and the graceful-stop event with its scope, is appended to the row's job summary.

## Cleanup

Cleanup runs twice against the same manifest: once inside `run` after the scenario ends, with its own three-minute allowance, and again from the workflow's `if: always()` step. Each pass:

- force-terminates what an agent the invocation owns still has running. On POSIX that is the agent's process group and the groups of its direct children. On Windows each child the harness starts runs in its own job object, which also holds every process it starts, including workers that outlive the agent and the agent's own nested worker job. Cleanup terminates the job and requires it to be empty, and the job is killed when the harness exits.
- resolves agents recorded by an earlier invocation. On POSIX a pid is signalled only while `ps` still reports the recorded start time and command line. On Windows cleanup holds a process handle across the check and terminates the process and its descendants only while that handle shows the recorded creation time and executable. A descendant counts only when it was created after its parent. Otherwise the pid is left alone, and a record from the other platform is never acted on. A record is marked reaped only once its process is confirmed gone. A pid that cannot be inspected, a foreign-platform record, or a killed process not confirmed gone within ten seconds fails cleanup and keeps its record for the next pass, while the controller steps below still run.
- unshares every share listed in this run's environment, once the environment's description confirms it belongs to this run.
- deletes this run's name; if a live share still holds it, retries after the environment is disabled.
- disables this run's environment.
- verifies with fresh successful reads that the environment, the name and every recorded token are absent.

When enable failed after the controller created the environment, cleanup finds it by its unique description. Two environments with that description, or a recorded environment whose description differs, are reported and left in place. Cleanup never removes resources by prefix or by account. A failed read leaves cleanup failed and lists the identifiers that remain. Cleanup success never turns a failed assertion into a passing run, and repeating cleanup after success makes no further changes.

## Recovery

Runner loss or forced cancellation can stop both cleanup passes; `always()` does not survive the runner's destruction. The harness logs each identifier as it acquires it (`recorded environment '<zId>'`, `recorded name '<namespace>:<name>'`, `recorded share '<token>'`, `recorded agent pid`), and the manifest's name and description appear in the log before the first side effect. To remove one run's leftovers, write a manifest with the logged values into an empty directory and run cleanup from any machine. Cleanup does not read `~/.zrok2`.

```json
{
  "run_id": "<run id>",
  "run_attempt": "<attempt>",
  "api_endpoint": "https://api-v2.zrok.io",
  "namespace": "public",
  "name": "agentsmoke-<run id>-<attempt>-<random>",
  "description": "zrok agent smoke <run id>-<attempt>-<random>",
  "name_attempted": true,
  "env_attempted": true,
  "env_zid": "<zId, if logged>",
  "share_tokens": ["<token>", "..."]
}
```

```
ZROK2_CI_ACCOUNT_TOKEN=<dedicated account token> agent-smoke cleanup --state-dir <dir>
```

Without a logged environment identity, cleanup matches the exact description. Anything it cannot attribute to the run is reported, not removed. Overlap between runs is reduced by the concurrency group; there is no scheduled janitor.

## Validation status

Hosted run [41112465896](https://github.com/openziti/zrok/actions/runs/41112465896) passed on 2026-10-30 at commit `ee5cbd392a34ced018d608699ee6cc6e0b838ece`, using the `macos-15` arm64 runner against the production zrok service. Native process/worker fixtures, agent lifecycle race tests, offline harness tests, the live scenario and fallback cleanup all passed. This supplies the macOS evidence for successful SIGINT shutdown, old-token absence, reserved-name preservation, automatic replay with a different token, public HTTP traffic and final resource cleanup. That run came from the earlier macOS-only workflow, `macos-agent-smoke.yml`, which this workflow replaces.

On 2026-10-30, the operator reported that all three rows of the generalized workflow passed on `fixes-209`: macOS arm64, Linux amd64 and Windows amd64. This supplies the first native Windows pass and regression passes for macOS and Linux after the shared harness refactor. The temporary push trigger has been removed. The run URL, tested commit and detailed step results have not yet been attached here; this result is operator-reported rather than independently inspected.
