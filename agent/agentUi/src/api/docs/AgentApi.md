# AgentagentGrpcagentproto.AgentApi

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**agentAccessPrivate**](AgentApi.md#agentAccessPrivate) | **POST** /v1/agent/accessPrivate | 
[**agentReleaseAccess**](AgentApi.md#agentReleaseAccess) | **POST** /v1/agent/releaseAccess | 
[**agentReleaseShare**](AgentApi.md#agentReleaseShare) | **POST** /v1/agent/releaseShare | 
[**agentSharePrivate**](AgentApi.md#agentSharePrivate) | **POST** /v1/agent/sharePrivate | 
[**agentSharePublic**](AgentApi.md#agentSharePublic) | **POST** /v1/agent/sharePublic | 
[**agentStatus**](AgentApi.md#agentStatus) | **GET** /v1/agent/status | 
[**agentVersion**](AgentApi.md#agentVersion) | **GET** /v1/agent/version | 

<a name="agentAccessPrivate"></a>
# **agentAccessPrivate**
> AccessPrivateResponse agentAccessPrivate(opts)



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
let opts = { 
  'token': "token_example", // String | 
  'bindAddress': "bindAddress_example", // String | 
  'responseHeaders': ["responseHeaders_example"] // [String] | 
};
apiInstance.agentAccessPrivate(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **String**|  | [optional] 
 **bindAddress** | **String**|  | [optional] 
 **responseHeaders** | [**[String]**](String.md)|  | [optional] 

### Return type

[**AccessPrivateResponse**](AccessPrivateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

<a name="agentReleaseAccess"></a>
# **agentReleaseAccess**
> ReleaseAccessResponse agentReleaseAccess(opts)



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
let opts = { 
  'frontendToken': "frontendToken_example" // String | 
};
apiInstance.agentReleaseAccess(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **frontendToken** | **String**|  | [optional] 

### Return type

[**ReleaseAccessResponse**](ReleaseAccessResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

<a name="agentReleaseShare"></a>
# **agentReleaseShare**
> ReleaseShareResponse agentReleaseShare(opts)



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
let opts = { 
  'token': "token_example" // String | 
};
apiInstance.agentReleaseShare(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **token** | **String**|  | [optional] 

### Return type

[**ReleaseShareResponse**](ReleaseShareResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

<a name="agentSharePrivate"></a>
# **agentSharePrivate**
> SharePrivateResponse agentSharePrivate(opts)



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
let opts = { 
  'target': "target_example", // String | 
  'backendMode': "backendMode_example", // String | 
  'insecure': true, // Boolean | 
  'closed': true, // Boolean | 
  'accessGrants': ["accessGrants_example"] // [String] | 
};
apiInstance.agentSharePrivate(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **target** | **String**|  | [optional] 
 **backendMode** | **String**|  | [optional] 
 **insecure** | **Boolean**|  | [optional] 
 **closed** | **Boolean**|  | [optional] 
 **accessGrants** | [**[String]**](String.md)|  | [optional] 

### Return type

[**SharePrivateResponse**](SharePrivateResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

<a name="agentSharePublic"></a>
# **agentSharePublic**
> SharePublicResponse agentSharePublic(opts)



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
let opts = { 
  'target': "target_example", // String | 
  'basicAuth': ["basicAuth_example"], // [String] | 
  'frontendSelection': ["frontendSelection_example"], // [String] | 
  'backendMode': "backendMode_example", // String | 
  'insecure': true, // Boolean | 
  'oauthProvider': "oauthProvider_example", // String | 
  'oauthEmailAddressPatterns': ["oauthEmailAddressPatterns_example"], // [String] | 
  'oauthCheckInterval': "oauthCheckInterval_example", // String | 
  'closed': true, // Boolean | 
  'accessGrants': ["accessGrants_example"] // [String] | 
};
apiInstance.agentSharePublic(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **target** | **String**|  | [optional] 
 **basicAuth** | [**[String]**](String.md)|  | [optional] 
 **frontendSelection** | [**[String]**](String.md)|  | [optional] 
 **backendMode** | **String**|  | [optional] 
 **insecure** | **Boolean**|  | [optional] 
 **oauthProvider** | **String**|  | [optional] 
 **oauthEmailAddressPatterns** | [**[String]**](String.md)|  | [optional] 
 **oauthCheckInterval** | **String**|  | [optional] 
 **closed** | **Boolean**|  | [optional] 
 **accessGrants** | [**[String]**](String.md)|  | [optional] 

### Return type

[**SharePublicResponse**](SharePublicResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

<a name="agentStatus"></a>
# **agentStatus**
> StatusResponse agentStatus()



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
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

<a name="agentVersion"></a>
# **agentVersion**
> VersionResponse agentVersion()



### Example
```javascript
import {AgentagentGrpcagentproto} from 'agentagent_grpcagentproto';

let apiInstance = new AgentagentGrpcagentproto.AgentApi();
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

