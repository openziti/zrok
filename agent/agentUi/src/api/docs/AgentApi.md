# AgentAgentGrpcAgentProto.AgentApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**agentStatus**](AgentApi.md#agentStatus) | **GET** /v1/agent/status | 
[**agentVersion**](AgentApi.md#agentVersion) | **GET** /v1/agent/version | 



## agentStatus

> StatusResponse agentStatus()



### Example

```javascript
import AgentAgentGrpcAgentProto from 'agent_agent_grpc_agent_proto';

let apiInstance = new AgentAgentGrpcAgentProto.AgentApi();
apiInstance.agentStatus((error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**StatusResponse**](StatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## agentVersion

> VersionResponse agentVersion()



### Example

```javascript
import AgentAgentGrpcAgentProto from 'agent_agent_grpc_agent_proto';

let apiInstance = new AgentAgentGrpcAgentProto.AgentApi();
apiInstance.agentVersion((error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**VersionResponse**](VersionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

