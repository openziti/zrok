# Zrok.AdminApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createFrontend**](AdminApi.md#createFrontend) | **POST** /frontend | 
[**createIdentity**](AdminApi.md#createIdentity) | **POST** /identity | 
[**deleteFrontend**](AdminApi.md#deleteFrontend) | **DELETE** /frontend | 
[**inviteTokenGenerate**](AdminApi.md#inviteTokenGenerate) | **POST** /invite/token/generate | 
[**listFrontends**](AdminApi.md#listFrontends) | **GET** /frontends | 
[**updateFrontend**](AdminApi.md#updateFrontend) | **PATCH** /frontend | 



## createFrontend

> CreateFrontendResponse createFrontend(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
let opts = {
  'body': new Zrok.CreateFrontendRequest() // CreateFrontendRequest | 
};
apiInstance.createFrontend(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateFrontendRequest**](CreateFrontendRequest.md)|  | [optional] 

### Return type

[**CreateFrontendResponse**](CreateFrontendResponse.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## createIdentity

> CreateIdentity201Response createIdentity(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
let opts = {
  'body': new Zrok.CreateIdentityRequest() // CreateIdentityRequest | 
};
apiInstance.createIdentity(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateIdentityRequest**](CreateIdentityRequest.md)|  | [optional] 

### Return type

[**CreateIdentity201Response**](CreateIdentity201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## deleteFrontend

> deleteFrontend(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
let opts = {
  'body': new Zrok.DeleteFrontendRequest() // DeleteFrontendRequest | 
};
apiInstance.deleteFrontend(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DeleteFrontendRequest**](DeleteFrontendRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined


## inviteTokenGenerate

> inviteTokenGenerate(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
let opts = {
  'body': new Zrok.InviteTokenGenerateRequest() // InviteTokenGenerateRequest | 
};
apiInstance.inviteTokenGenerate(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**InviteTokenGenerateRequest**](InviteTokenGenerateRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined


## listFrontends

> [PublicFrontend] listFrontends()



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
apiInstance.listFrontends().then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters

This endpoint does not need any parameter.

### Return type

[**[PublicFrontend]**](PublicFrontend.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/zrok.v1+json


## updateFrontend

> updateFrontend(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.AdminApi();
let opts = {
  'body': new Zrok.UpdateFrontendRequest() // UpdateFrontendRequest | 
};
apiInstance.updateFrontend(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateFrontendRequest**](UpdateFrontendRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined

