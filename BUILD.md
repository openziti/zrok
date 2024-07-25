# Build

## zrok

At this time, building `zrok` is pretty straightforward. You will require `node` v18+ to be installed in order to complete the build as well as `go`. Because `zrok` uses CGO, you will also need to have a working C compiler toolchain. [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/) works great on Windows (just make sure it's in your PATH).

To build, follow these steps:

* clone the repository
* change to the existing `ui` folder
* run `npm install`
* run `npm run build` (this process takes a while the first time and only needs to be run if the ui changes)
* change back to the checkout root
* make sure the dist directory exists: `mkdir -p dist`
* build the go project normally: `go build -o dist ./...`

## Cross-build zrok with Docker

Build a `zrok` snapshot binary for amd64, arm64, armhf, or armel with [the `cross-build` container](./docker/images/cross-build/README.md).

## Documentation/Website

The doc website is based on [Docusaurus](https://docusaurus.io/) which in turn will require `npm` to be installed. `yarn`
is another tool which is used to start the Docusaurus dev site.

To build the doc:

* cd to `website`
* run `yarn install` (usually only needed once)
* run `yarn start` to start the development server (make sure port 3000 is open or change the port)

The development server infrequently behaves differently than the 'production' build. If you must use the 'production'
build it is slower, but you can accomplish that with `yarn build`.
