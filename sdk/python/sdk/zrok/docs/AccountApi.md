# zrok_api.AccountApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**change_password**](AccountApi.md#change_password) | **POST** /changePassword | 
[**invite**](AccountApi.md#invite) | **POST** /invite | 
[**login**](AccountApi.md#login) | **POST** /login | 
[**regenerate_account_token**](AccountApi.md#regenerate_account_token) | **POST** /regenerateAccountToken | 
[**register**](AccountApi.md#register) | **POST** /register | 
[**reset_password**](AccountApi.md#reset_password) | **POST** /resetPassword | 
[**reset_password_request**](AccountApi.md#reset_password_request) | **POST** /resetPasswordRequest | 
[**verify**](AccountApi.md#verify) | **POST** /verify | 


# **change_password**
> change_password(body=body)



### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.change_password_request import ChangePasswordRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: key
configuration.api_key['key'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['key'] = 'Bearer'

# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.ChangePasswordRequest() # ChangePasswordRequest |  (optional)

    try:
        api_instance.change_password(body=body)
    except Exception as e:
        print("Exception when calling AccountApi->change_password: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ChangePasswordRequest**](ChangePasswordRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | password changed |  -  |
**400** | password not changed |  -  |
**401** | unauthorized |  -  |
**422** | password validation failure |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **invite**
> invite(body=body)



### Example


```python
import zrok_api
from zrok_api.models.invite_request import InviteRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.InviteRequest() # InviteRequest |  (optional)

    try:
        api_instance.invite(body=body)
    except Exception as e:
        print("Exception when calling AccountApi->invite: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**InviteRequest**](InviteRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | invitation created |  -  |
**400** | invitation not created (already exists) |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> str login(body=body)



### Example


```python
import zrok_api
from zrok_api.models.login_request import LoginRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.LoginRequest() # LoginRequest |  (optional)

    try:
        api_response = api_instance.login(body=body)
        print("The response of AccountApi->login:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AccountApi->login: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**LoginRequest**](LoginRequest.md)|  | [optional] 

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | login successful |  -  |
**401** | invalid login |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **regenerate_account_token**
> RegenerateAccountToken200Response regenerate_account_token(body=body)



### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.regenerate_account_token200_response import RegenerateAccountToken200Response
from zrok_api.models.regenerate_account_token_request import RegenerateAccountTokenRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: key
configuration.api_key['key'] = os.environ["API_KEY"]

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['key'] = 'Bearer'

# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.RegenerateAccountTokenRequest() # RegenerateAccountTokenRequest |  (optional)

    try:
        api_response = api_instance.regenerate_account_token(body=body)
        print("The response of AccountApi->regenerate_account_token:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AccountApi->regenerate_account_token: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegenerateAccountTokenRequest**](RegenerateAccountTokenRequest.md)|  | [optional] 

### Return type

[**RegenerateAccountToken200Response**](RegenerateAccountToken200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | regenerate account token |  -  |
**404** | account not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **register**
> RegenerateAccountToken200Response register(body=body)



### Example


```python
import zrok_api
from zrok_api.models.regenerate_account_token200_response import RegenerateAccountToken200Response
from zrok_api.models.register_request import RegisterRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.RegisterRequest() # RegisterRequest |  (optional)

    try:
        api_response = api_instance.register(body=body)
        print("The response of AccountApi->register:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AccountApi->register: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegisterRequest**](RegisterRequest.md)|  | [optional] 

### Return type

[**RegenerateAccountToken200Response**](RegenerateAccountToken200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | account created |  -  |
**404** | request not found |  -  |
**422** | password validation failure |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **reset_password**
> reset_password(body=body)



### Example


```python
import zrok_api
from zrok_api.models.reset_password_request import ResetPasswordRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.ResetPasswordRequest() # ResetPasswordRequest |  (optional)

    try:
        api_instance.reset_password(body=body)
    except Exception as e:
        print("Exception when calling AccountApi->reset_password: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ResetPasswordRequest**](ResetPasswordRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | password reset |  -  |
**404** | request not found |  -  |
**422** | password validation failure |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **reset_password_request**
> reset_password_request(body=body)



### Example


```python
import zrok_api
from zrok_api.models.regenerate_account_token_request import RegenerateAccountTokenRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.RegenerateAccountTokenRequest() # RegenerateAccountTokenRequest |  (optional)

    try:
        api_instance.reset_password_request(body=body)
    except Exception as e:
        print("Exception when calling AccountApi->reset_password_request: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegenerateAccountTokenRequest**](RegenerateAccountTokenRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | reset password request created |  -  |
**400** | reset password request not created |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **verify**
> Verify200Response verify(body=body)



### Example


```python
import zrok_api
from zrok_api.models.verify200_response import Verify200Response
from zrok_api.models.verify_request import VerifyRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v1"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.AccountApi(api_client)
    body = zrok_api.VerifyRequest() # VerifyRequest |  (optional)

    try:
        api_response = api_instance.verify(body=body)
        print("The response of AccountApi->verify:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AccountApi->verify: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**VerifyRequest**](VerifyRequest.md)|  | [optional] 

### Return type

[**Verify200Response**](Verify200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | registration token ready |  -  |
**404** | registration token not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

