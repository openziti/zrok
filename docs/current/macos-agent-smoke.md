# macOS agent smoke

The `macOS agent smoke` workflow (`.github/workflows/macos-agent-smoke.yml`) runs the agent's named public share through an interrupt and a restart on a native GitHub-hosted macOS runner, against a live zrok v2 service with a dedicated test account. It builds `zrok2` from the triggering commit and drives it as a user would. It supplies the native evidence that [Agent share lifecycle](agent-share-lifecycle.md) records as outstanding for macOS. It is the only CI job that uses a live service account. It runs on manual dispatch and, temporarily, pushes to `fixes-209` for validation before merging. No other workflow, required check, release or `Makefile` target depends on it.

## What a run proves

The live scenario passes only when every observation holds and final cleanup succeeds:

1. A fresh environment is enabled with the built CLI's headless enable path, and a new unique name is reserved.
2. `zrok2 agent start` runs in its own process group and reports the exact build identity the job stamped.
3. `zrok2 share public --force-agent --open -n <namespace>:<name>` against a loopback server produces exactly one active share in structured agent status, the name is bound to its token, and the public endpoint returns the server's per-run random body. Bare endpoint hostnames are fetched over HTTPS; explicit HTTP or HTTPS schemes are preserved.
4. SIGINT to the agent's process group, as a terminal Ctrl-C delivers it, ends the agent with exit 0 within 130 seconds. A forced kill is a failure.
5. A successful authenticated account overview still lists the environment, no longer mentions the share token anywhere, and the name listing shows the name reserved and unbound.
6. Restarting the agent with the same environment replays the saved request without another share command: exactly one active share, a new token, the old token still absent, the name bound to the replacement and the random body served through it.
7. A second interrupt repeats the exit, absence and name checks. Cleanup then removes the name and the environment and proves their absence.

Before the live scenario the job runs the native process and worker fixtures (`./agent/proctree`, `./cmd/zrok2/subordinate`), the agent package's lifecycle tests under the race detector, and the harness's offline tests.

A successful build, passing native fixtures or a run rescued by fallback cleanup does not close the macOS gap; only a run whose scenario steps and run-mode cleanup all pass does. Windows validation and the Linux live check are outside this workflow and remain outstanding.

## Setup

The operator does these once, in repository settings. No step needs an admin token.

- A dedicated v2 account on the target service, used for nothing else, permitted to enable an environment and reserve a name in the target namespace. The job never provisions or modifies the account itself.
- A GitHub environment named `zrok-agent-smoke`, with deployment branches restricted to trusted branches, holding the secret `ZROK2_CI_ACCOUNT_TOKEN` (the dedicated account's token). An environment secret reaches only jobs that reference that environment, and within the job only the live scenario and fallback cleanup steps receive it.
- Optionally, environment variables `ZROK2_SMOKE_API_ENDPOINT` and `ZROK2_SMOKE_NAMESPACE` to target a dedicated service. The defaults are `https://api-v2.zrok.io` and `public`.

A missing secret fails the live step with an error naming it; the job is never green without the live scenario.

## Dispatch

To validate before merging, the workflow temporarily also triggers on every push to `fixes-209`. Commit and push the workflow change on that branch, and ensure the `zrok-agent-smoke` environment admits `fixes-209`. GitHub reads the workflow from the pushed commit, so this first run needs no change on `main`. Find it in Actions under the pushed commit, or use `gh run list --branch fixes-209 --workflow macos-agent-smoke.yml`. Remove the temporary `push` trigger after branch validation and before merging to restore manual-only operation.

GitHub's manual Run workflow button requires the workflow on the default branch. Once it reaches `main`, an authorized operator selects the branch to test: Actions, `macOS agent smoke`, Run workflow, or `gh workflow run macos-agent-smoke.yml --ref <branch>`. The branch must be one the `zrok-agent-smoke` environment admits. The workflow takes no inputs: the checkout is the triggering ref, and the endpoint and namespace come only from the environment's variables.

Runs share one concurrency group and never cancel each other; a second dispatch waits for the first. The job is bounded at 45 minutes, the live step at 20 and fallback cleanup at 5.

The build version is the latest `v2.*` tag reachable from the checkout, and `build.Hash` is the full commit hash. Both are written to the job summary. The CLI and the harness are built with identical stamps, so the controller's version check applies to both. A version refusal fails the run as a failed prerequisite and keeps the controller's message.

## The harness

`tests/agent-smoke` is a Go command whose sources all carry the `agent_smoke` build tag, plus `unix`. Ordinary `go build ./...`, `go test ./...`, `go vet ./...`, `go install ./...` and `make test` never select it. Build it with `go build -tags agent_smoke ./tests/agent-smoke`.

```
agent-smoke run --binary <zrok2> --state-dir <dir> [--api-endpoint <url>] [--namespace <token>]
agent-smoke cleanup --state-dir <dir>
```

Both modes read the account token from `ZROK2_CI_ACCOUNT_TOKEN`. Only `run` creates resources. It refuses before any side effect when `~/.zrok2` exists, when the state directory already holds a manifest, or when the harness was built without `build.Version` and `build.Hash`. The CLI it drives always uses `~/.zrok2`, so a live run needs a fresh home: the hosted runner's, or a disposable OS user's when run locally. The harness never adopts, modifies or cleans an existing environment. The offline tests run anywhere: `go test -tags agent_smoke ./tests/agent-smoke`. They use a fake controller, a fake public frontend and the test binary re-executed as a fake `zrok2`, and never contact a zrok service.

Before any side effect, `run` chooses the name `agentsmoke-<run id>-<attempt>-<random>` and the environment description `zrok agent smoke <run id>-<attempt>-<random>`, from `GITHUB_RUN_ID` and `GITHUB_RUN_ATTEMPT` or `local-0` outside Actions. It records them in `<state-dir>/manifest.json` (mode 0600). It then adds each identifier as it is acquired: the environment's Ziti identity, every share token observed and each agent's pid and start time. The manifest never holds the account token or identity material. A name already listed for the account fails the run, and that name is marked foreign so cleanup never touches it.

The scenario has an overall 15-minute bound. Agent readiness and share startup or replay each get at most 7 minutes, an agent stop 130 seconds and public HTTP convergence 120 seconds. Each controller or agent request gets 30 seconds, and each short CLI invocation 90 seconds. All of these are capped by the scenario's remaining time. Readiness is polled, never inferred from a sleep. The harness's own assertions read the generated REST client and the agent's structured gRPC status, independently of the agent's cleanup code.

Child output is held in memory (the first 64 KiB and last 192 KiB per process) and written to `<state-dir>/report/logs/` only after the account token has been redacted. The report directory also holds `run-summary.json`/`.md` and `cleanup-summary.json`/`.md`. It is the only uploaded artifact, kept for seven days, and its markdown summary is appended to the job summary.

## Cleanup

Cleanup runs twice against the same manifest: once inside `run` after the scenario ends, with its own three-minute allowance, and again from the workflow's `if: always()` step. Each pass:

- force-terminates an agent the invocation owns. A pid recorded by an earlier invocation is signalled only while `ps` still reports the recorded start time and command line; otherwise it is left alone. A record is marked reaped only once its process is confirmed gone. A pid that cannot be inspected, or a killed process not confirmed gone within ten seconds, fails cleanup and keeps its record for the next pass, while the controller steps below still run.
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

Hosted run [41110212451](https://github.com/openziti/zrok/actions/runs/41110212451), at commit `9b6bdf90eb01be04a894e33f196c9903d03da55d` on the macOS arm64 job, reached the live scenario and created a named share. The public response check failed because the harness treated the returned hostname as a URL without a scheme. Fallback cleanup confirmed removal of the share, name and environment. The harness now defaults bare endpoints to HTTPS and awaits another hosted run; neither graceful shutdown nor replay was reached in that attempt.

No passing hosted live scenario has been recorded. Local Linux/amd64 verification includes tagged offline tests under the race detector and ordinary package discovery excluding the harness. The workflow file was parsed but not checked with `actionlint`, which was unavailable. When a hosted run completes, record its URL, commit, runner OS and architecture, scenario result and cleanup outcome here. Only a passing run updates the macOS line in [Agent share lifecycle](agent-share-lifecycle.md).
