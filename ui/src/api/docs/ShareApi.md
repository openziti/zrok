# Zrok.ShareApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**access**](ShareApi.md#access) | **POST** /access | 
[**share**](ShareApi.md#share) | **POST** /share | 
[**unaccess**](ShareApi.md#unaccess) | **DELETE** /unaccess | 
[**unshare**](ShareApi.md#unshare) | **DELETE** /unshare | 
[**updateShare**](ShareApi.md#updateShare) | **PATCH** /share | 



## access

> AccessResponse access(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.ShareApi();
let opts = {
  'body': new Zrok.AccessRequest() // AccessRequest | 
};
apiInstance.access(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AccessRequest**](AccessRequest.md)|  | [optional] 

### Return type

[**AccessResponse**](AccessResponse.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## share

> ShareResponse share(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.ShareApi();
let opts = {
  'body': new Zrok.ShareRequest() // ShareRequest | 
};
apiInstance.share(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ShareRequest**](ShareRequest.md)|  | [optional] 

### Return type

[**ShareResponse**](ShareResponse.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## unaccess

> unaccess(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.ShareApi();
let opts = {
  'body': new Zrok.UnaccessRequest() // UnaccessRequest | 
};
apiInstance.unaccess(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnaccessRequest**](UnaccessRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined


## unshare

> unshare(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.ShareApi();
let opts = {
  'body': new Zrok.UnshareRequest() // UnshareRequest | 
};
apiInstance.unshare(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnshareRequest**](UnshareRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## updateShare

> updateShare(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.ShareApi();
let opts = {
  'body': new Zrok.UpdateShareRequest() // UpdateShareRequest | 
};
apiInstance.updateShare(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateShareRequest**](UpdateShareRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined

