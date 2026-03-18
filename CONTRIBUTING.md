# Contributing

NetFoundry welcomes all and any contributions. All open source projects managed by NetFoundry share a common
[guide for contributions](https://netfoundry.io/docs/openziti/policies/CONTRIBUTING/).

If you are eager to contribute to a NetFoundry-managed open source project please read and act accordingly.

## AI coding agents

If you use an AI coding agent (Claude Code, Cursor, Copilot, etc.), configure it to follow the conventions in this
file. The project's `CLAUDE.md` is git-ignored and will not be present in a fresh clone, so your local agent
configuration must reference CONTRIBUTING.md as the authoritative source of project guidelines.

## Naming: zrok vs `zrok2`

+ **zrok** (lowercase, never code-quoted) is the project's proper name used in prose, headings, and branding. Never write "Zrok", "ZROK", or `` `zrok` `` when referring to the project.
+ **`zrok2`** (code-quoted) is the CLI binary name. Always use backtick quoting when referring to the executable, e.g., "run `zrok2 share public`."
+ Other code-related occurrences of the `zrok2` namespace — package names, distribution artifacts, configuration keys, service units, environment variables, etc. — should also be code-quoted because they are literal identifiers (e.g., `zrok2-share`, `zrok2.service`, `ZROK2_TOKEN`).

## Project style guides

This project uses several programming languages. Each language has its own style guide specifying any language-specific
conventions and idioms. The log formatting examples for Go are applicable to all languages.

### Markdown

+ This project uses [GitHub Flavored Markdown](https://github.github.com/gfm/).
+ Wrap lines at 120 characters if the audience is expected to read the source.
+ Do not wrap lines if the audience is expected to read the rendered output as HTML.
+ Use [Markdownlint](https://github.com/DavidAnson/markdownlint) with [the configuration file](.markdownlint.yaml) committed to this repository to find formatting problems. Run `npx markdownlint-cli <file>` on every changed Markdown file before submitting a pull request.
+ Use sentence-style capitalization for all Markdown headings in every file (README, CONTRIBUTING, docs, etc.): capitalize only the first letter of the heading (and any proper nouns or acronyms). Do not use title case.
  + Correct: `## Quick start`, `### Configure OpenZiti metrics events`
  + Incorrect: `## Quick Start`, `### Configure OpenZiti Metrics Events`

### Go

+ This project uses [Go](https://golang.org/) language conventions.
+ Organize imports as a single block in alphabetical order without empty lines.
+ Begin log messages with a lowercase letter and do not end with punctuation. This log formatting guidance applies to all languages, not only Go.

    Format log messages that report errors.

    ```go
    dl.Errorf("tried a thing and failed: %v", err)
    ```

    Format in-line information in informational log messages.

    ```go
    dl.Infof("the expected value '%v' arrived as '%v'", expected, actual)
    ```

    Format in-line information in error log messages.

    ```go
    dl.Errorf("the expected value '%v did not compute: %v", value, err)
    ```

+ Format log messages with format strings and arguments like 'tried a thing and failed: %s'.

### Python

+ This project uses [Python](https://www.python.org/) language conventions PEP-8.
+ Use [flake8](https://flake8.pycqa.org/en/latest/) with [the configuration file](.flake8) committed to this repository to find formatting problems.
+ Use Go log formatting guidance for Python too.

### Docusaurus

+ This project uses [Docusaurus](https://docusaurus.io/) with NodeJS 20+ to build static content for docs.zrok.io.
+ Use `npm` to manage Node modules, not `yarn` (Ken plans to switch from `npm` to `yarn` if no one else does).
+ Documentation changes must always target the current (unversioned) docs tree at `website/docs/`. The `versioned_docs/` and `zrok_versioned_docs/` directories preserve past generations of documentation and are not maintained.
+ Follow the sentence-case heading rule described in the Markdown section above.
