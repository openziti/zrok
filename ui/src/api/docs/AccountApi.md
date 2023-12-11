# Zrok.AccountApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**invite**](AccountApi.md#invite) | **POST** /invite | 
[**login**](AccountApi.md#login) | **POST** /login | 
[**register**](AccountApi.md#register) | **POST** /register | 
[**resetPassword**](AccountApi.md#resetPassword) | **POST** /resetPassword | 
[**resetPasswordRequest**](AccountApi.md#resetPasswordRequest) | **POST** /resetPasswordRequest | 
[**verify**](AccountApi.md#verify) | **POST** /verify | 



## invite

> invite(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.InviteRequest() // InviteRequest | 
};
apiInstance.invite(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**InviteRequest**](InviteRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## login

> String login(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.LoginRequest() // LoginRequest | 
};
apiInstance.login(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**LoginRequest**](LoginRequest.md)|  | [optional] 

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## register

> RegisterResponse register(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.RegisterRequest() // RegisterRequest | 
};
apiInstance.register(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegisterRequest**](RegisterRequest.md)|  | [optional] 

### Return type

[**RegisterResponse**](RegisterResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## resetPassword

> resetPassword(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.ResetPasswordRequest() // ResetPasswordRequest | 
};
apiInstance.resetPassword(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ResetPasswordRequest**](ResetPasswordRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json


## resetPasswordRequest

> resetPasswordRequest(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.ResetPasswordRequestRequest() // ResetPasswordRequestRequest | 
};
apiInstance.resetPasswordRequest(opts).then(() => {
  console.log('API called successfully.');
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ResetPasswordRequestRequest**](ResetPasswordRequestRequest.md)|  | [optional] 

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: Not defined


## verify

> VerifyResponse verify(opts)



### Example

```javascript
import Zrok from 'zrok';

let apiInstance = new Zrok.AccountApi();
let opts = {
  'body': new Zrok.VerifyRequest() // VerifyRequest | 
};
apiInstance.verify(opts).then((data) => {
  console.log('API called successfully. Returned data: ' + data);
}, (error) => {
  console.error(error);
});

```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**VerifyRequest**](VerifyRequest.md)|  | [optional] 

### Return type

[**VerifyResponse**](VerifyResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/zrok.v1+json
- **Accept**: application/zrok.v1+json

