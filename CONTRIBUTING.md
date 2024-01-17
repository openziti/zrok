# Contributing

NetFoundry welcomes all and any contributions. All open source projects managed by NetFoundry share a common
[guide for contributions](https://netfoundry.github.io/policies/CONTRIBUTING.html).

If you are eager to contribute to a NetFoundry-managed open source project please read and act accordingly.

## Project Styles

This project uses several programming languages. Each language has its own style guide specifying any language-specific
conventions and idioms. The log formatting examples for Go are applicable to all languages.

### Markdown

+ This project uses [GitHub Flavored Markdown](https://github.github.com/gfm/).
+ Wrap lines at 120 characters if the audience is expected to read the source.
+ Do not wrap lines if the audience is expected to read the rendered output as HTML.
+ Use [Markdownlint](https://github.com/DavidAnson/markdownlint) with [the configuration file](.markdownlint.yaml) committed to this repository to find formatting problems.

### Go

+ This project uses [Go](https://golang.org/) language conventions.
+ Organize imports as a single block in alphabetical order without empty lines.
+ Begin log messages with a lowercase letter and do not end with punctuation. This log formatting guidance applies to all languages, not only Go.

    Format log messages that report errors.

    ```go
    logrus.Errorf("tried a thing and failed: %v", err)
    ```

    Format in-line information in informational log messages.

    ```go
    logrus.Infof("the expected value '%v' arrived as '%v'", expected, actual)
    ```

    Format in-line information in error log messages.

    ```go
    logrus.Errorf("the expected value '%v did not compute: %v", value, err)
    ```

+ Format log messages with format strings and arguments like 'tried a thing and failed: %s'.

### Python

+ This project uses [Python](https://www.python.org/) language conventions PEP-8.
+ Use [flake8](https://flake8.pycqa.org/en/latest/) with [the configuration file](.flake8) committed to this repository to find formatting problems.
+ Use Go log formatting guidance for Python too.

### Docusaurus

+ This project uses [Docusaurus](https://docusaurus.io/) with NodeJS 18 to build static content for docs.zrok.io.
+ Use `npm` to manage Node modules, not `yarn` (Ken plans to switch from `npm` to `yarn` if no one else does).
