# Zrok.MetadataApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**configuration**](MetadataApi.md#configuration) | **GET** /configuration | 
[**getAccountDetail**](MetadataApi.md#getAccountDetail) | **GET** /detail/account | 
[**getAccountMetrics**](MetadataApi.md#getAccountMetrics) | **GET** /metrics/account | 
[**getEnvironmentDetail**](MetadataApi.md#getEnvironmentDetail) | **GET** /detail/environment/{envZId} | 
[**getEnvironmentMetrics**](MetadataApi.md#getEnvironmentMetrics) | **GET** /metrics/environment/{envId} | 
[**getFrontendDetail**](MetadataApi.md#getFrontendDetail) | **GET** /detail/frontend/{feId} | 
[**getShareDetail**](MetadataApi.md#getShareDetail) | **GET** /detail/share/{shrToken} | 
[**getShareMetrics**](MetadataApi.md#getShareMetrics) | **GET** /metrics/share/{shrToken} | 
[**overview**](MetadataApi.md#overview) | **GET** /overview | 
[**version**](MetadataApi.md#version) | **GET** /version | 



## configuration

> Configuration configuration()



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.MetadataApi();
apiInstance.configuration().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Configuration**](Configuration.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getAccountDetail

> [Environment] getAccountDetail()



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
apiInstance.getAccountDetail().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**[Environment]**](Environment.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getAccountMetrics

> Metrics getAccountMetrics(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let opts = {
  'duration': "duration_example" // String | 
};
apiInstance.getAccountMetrics(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **duration** | **String**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getEnvironmentDetail

> EnvironmentAndResources getEnvironmentDetail(envZId)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let envZId = "envZId_example"; // String | 
apiInstance.getEnvironmentDetail(envZId).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **envZId** | **String**|  | 

### Return type

[**EnvironmentAndResources**](EnvironmentAndResources.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getEnvironmentMetrics

> Metrics getEnvironmentMetrics(envId, opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let envId = "envId_example"; // String | 
let opts = {
  'duration': "duration_example" // String | 
};
apiInstance.getEnvironmentMetrics(envId, opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **envId** | **String**|  | 
 **duration** | **String**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getFrontendDetail

> Frontend getFrontendDetail(feId)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let feId = 56; // Number | 
apiInstance.getFrontendDetail(feId).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **feId** | **Number**|  | 

### Return type

[**Frontend**](Frontend.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getShareDetail

> Share getShareDetail(shrToken)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let shrToken = "shrToken_example"; // String | 
apiInstance.getShareDetail(shrToken).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **shrToken** | **String**|  | 

### Return type

[**Share**](Share.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## getShareMetrics

> Metrics getShareMetrics(shrToken, opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
let shrToken = "shrToken_example"; // String | 
let opts = {
  'duration': "duration_example" // String | 
};
apiInstance.getShareMetrics(shrToken, opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **shrToken** | **String**|  | 
 **duration** | **String**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## overview

> Overview overview()



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.MetadataApi();
apiInstance.overview().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Overview**](Overview.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## version

> String version()



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.MetadataApi();
apiInstance.version().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json

