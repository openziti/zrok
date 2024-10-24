# agentagent_grpcagentproto

AgentagentGrpcagentproto - JavaScript client for agentagent_grpcagentproto
No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
This SDK is automatically generated by the [Swagger Codegen](https://github.com/swagger-api/swagger-codegen) project:

- API version: version not set
- Package version: version not set
- Build package: io.swagger.codegen.v3.generators.javascript.JavaScriptClientCodegen

## Installation

### For [Node.js](https://nodejs.org/)

#### npm

To publish the library as a [npm](https://www.npmjs.com/),
please follow the procedure in ["Publishing npm packages"](https://docs.npmjs.com/getting-started/publishing-npm-packages).

Then install it via:

```shell
npm install agentagent_grpcagentproto --save
```

#### git
#
If the library is hosted at a git repository, e.g.
https://github.com/GIT_USER_ID/GIT_REPO_ID
then install it via:

```shell
    npm install GIT_USER_ID/GIT_REPO_ID --save
```

### For browser

The library also works in the browser environment via npm and [browserify](http://browserify.org/). After following
the above steps with Node.js and installing browserify with `npm install -g browserify`,
perform the following (assuming *main.js* is your entry file):

```shell
browserify main.js > bundle.js
```

Then include *bundle.js* in the HTML pages.

### Webpack Configuration

Using Webpack you may encounter the following error: "Module not found: Error:
Cannot resolve module", most certainly you should disable AMD loader. Add/merge
the following section to your webpack config:

```javascript
module: {
  rules: [
    {
      parser: {
        amd: false
      }
    }
  ]
}
```

## Getting Started

Please follow the [installation](#installation) instruction and execute the following JS code:

```javascript
var AgentagentGrpcagentproto = require('agentagent_grpcagentproto');

var api = new AgentagentGrpcagentproto.AgentApi()
var opts = { 
  'token': "token_example", // {String} 
  'bindAddress': "bindAddress_example", // {String} 
  'responseHeaders': ["responseHeaders_example"] // {[String]} 
};
var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
api.agentAccessPrivate(opts, callback);
```

## Documentation for API Endpoints

All URIs are relative to */*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AgentagentGrpcagentproto.AgentApi* | [**agentAccessPrivate**](docs/AgentApi.md#agentAccessPrivate) | **POST** /v1/agent/accessPrivate | 
*AgentagentGrpcagentproto.AgentApi* | [**agentReleaseAccess**](docs/AgentApi.md#agentReleaseAccess) | **POST** /v1/agent/releaseAccess | 
*AgentagentGrpcagentproto.AgentApi* | [**agentReleaseShare**](docs/AgentApi.md#agentReleaseShare) | **POST** /v1/agent/releaseShare | 
*AgentagentGrpcagentproto.AgentApi* | [**agentSharePrivate**](docs/AgentApi.md#agentSharePrivate) | **POST** /v1/agent/sharePrivate | 
*AgentagentGrpcagentproto.AgentApi* | [**agentSharePublic**](docs/AgentApi.md#agentSharePublic) | **POST** /v1/agent/sharePublic | 
*AgentagentGrpcagentproto.AgentApi* | [**agentStatus**](docs/AgentApi.md#agentStatus) | **GET** /v1/agent/status | 
*AgentagentGrpcagentproto.AgentApi* | [**agentVersion**](docs/AgentApi.md#agentVersion) | **GET** /v1/agent/version | 

## Documentation for Models

 - [AgentagentGrpcagentproto.AccessDetail](docs/AccessDetail.md)
 - [AgentagentGrpcagentproto.AccessPrivateResponse](docs/AccessPrivateResponse.md)
 - [AgentagentGrpcagentproto.ProtobufAny](docs/ProtobufAny.md)
 - [AgentagentGrpcagentproto.ReleaseAccessResponse](docs/ReleaseAccessResponse.md)
 - [AgentagentGrpcagentproto.ReleaseShareResponse](docs/ReleaseShareResponse.md)
 - [AgentagentGrpcagentproto.RpcStatus](docs/RpcStatus.md)
 - [AgentagentGrpcagentproto.ShareDetail](docs/ShareDetail.md)
 - [AgentagentGrpcagentproto.SharePrivateResponse](docs/SharePrivateResponse.md)
 - [AgentagentGrpcagentproto.SharePublicResponse](docs/SharePublicResponse.md)
 - [AgentagentGrpcagentproto.ShareReservedResponse](docs/ShareReservedResponse.md)
 - [AgentagentGrpcagentproto.StatusResponse](docs/StatusResponse.md)
 - [AgentagentGrpcagentproto.VersionResponse](docs/VersionResponse.md)

## Documentation for Authorization

 All endpoints do not require authorization.
