# Website

This website is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

### Installation

```
$ yarn
```

### Local Development

```
$ yarn start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

### Build

```
$ yarn build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.


### Cutting a new doc version

New doc releases should only be cut when major revisions are coming and the current version is ready to be frozen.
Cutting a new version will snapshot the current ./docs directory and copy it all into the ./website/versioned_docs directory based on the version that is tagged.

```
$ yarn docusaurus docs:version 1.1`
```

The default doc version that is displayed is managed in the `docusaurus.config.js` file.
By default the last version that was cut will be displayed, but this can be overridden be updating the config to render
the "current" doc version.

```
  presets: [
    [
        docs: {
          // These lines to show the current docs by default and assign them a label
          lastVersion: 'current',
          versions: {
             current: {
               label: '1.0',
             },
          },

        },
```

### Deployment

Using SSH:

```
$ USE_SSH=true yarn deploy
```

Not using SSH:

```
$ GIT_USER=<Your GitHub username> yarn deploy
```

If you are using GitHub pages for hosting, this command is a convenient way to build the website and push to the `gh-pages` branch.
