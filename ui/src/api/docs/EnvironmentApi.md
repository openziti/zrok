# Zrok.EnvironmentApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**disable**](EnvironmentApi.md#disable) | **POST** /disable | 
[**enable**](EnvironmentApi.md#enable) | **POST** /enable | 



## disable

> disable(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.EnvironmentApi();
let opts = {
  'body': new Zrok.DisableRequest() // DisableRequest | 
};
apiInstance.disable(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DisableRequest**](DisableRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined


## enable

> EnableResponse enable(opts)



### Example

```javascript
import Zrok from 'zrok';
let defaultClient = Zrok.ApiClient.instance;
// Configure API key authorization: key
let key = defaultClient.authentications['key'];
key.apiKey = 'YOUR API KEY';
// Uncomment the following line to set a prefix for the API key, e.g. "Token" (defaults to null)
//key.apiKeyPrefix = 'Token';

let apiInstance = new Zrok.EnvironmentApi();
let opts = {
  'body': new Zrok.EnableRequest() // EnableRequest | 
};
apiInstance.enable(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnableRequest**](EnableRequest.md)|  | [optional] 

### Return type

[**EnableResponse**](EnableResponse.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json

