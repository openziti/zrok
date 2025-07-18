# zrok_api.AgentApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**enroll**](AgentApi.md#enroll) | **POST** /agent/enroll | 
[**ping**](AgentApi.md#ping) | **POST** /agent/ping | 
[**remote_access**](AgentApi.md#remote_access) | **POST** /agent/access | 
[**remote_share**](AgentApi.md#remote_share) | **POST** /agent/share | 
[**remote_status**](AgentApi.md#remote_status) | **POST** /agent/status | 
[**remote_unaccess**](AgentApi.md#remote_unaccess) | **POST** /agent/unaccess | 
[**remote_unshare**](AgentApi.md#remote_unshare) | **POST** /agent/unshare | 
[**share_http_healthcheck**](AgentApi.md#share_http_healthcheck) | **POST** /agent/share/http-healthcheck | 
[**unenroll**](AgentApi.md#unenroll) | **POST** /agent/unenroll | 


# **enroll**
> Enroll200Response enroll(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.enroll200_response import Enroll200Response
from zrok_api.models.enroll_request import EnrollRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.EnrollRequest() # EnrollRequest |  (optional)

    try:
        api_response = api_instance.enroll(body=body)
        print("The response of AgentApi->enroll:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->enroll: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnrollRequest**](EnrollRequest.md)|  | [optional] 

### Return type

[**Enroll200Response**](Enroll200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**400** | bad request; already enrolled |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ping**
> Ping200Response ping(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.enroll_request import EnrollRequest
from zrok_api.models.ping200_response import Ping200Response
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.EnrollRequest() # EnrollRequest |  (optional)

    try:
        api_response = api_instance.ping(body=body)
        print("The response of AgentApi->ping:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->ping: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnrollRequest**](EnrollRequest.md)|  | [optional] 

### Return type

[**Ping200Response**](Ping200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remote_access**
> CreateFrontend201Response remote_access(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_frontend201_response import CreateFrontend201Response
from zrok_api.models.remote_access_request import RemoteAccessRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.RemoteAccessRequest() # RemoteAccessRequest |  (optional)

    try:
        api_response = api_instance.remote_access(body=body)
        print("The response of AgentApi->remote_access:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->remote_access: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RemoteAccessRequest**](RemoteAccessRequest.md)|  | [optional] 

### Return type

[**CreateFrontend201Response**](CreateFrontend201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remote_share**
> RemoteShare200Response remote_share(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.remote_share200_response import RemoteShare200Response
from zrok_api.models.remote_share_request import RemoteShareRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.RemoteShareRequest() # RemoteShareRequest |  (optional)

    try:
        api_response = api_instance.remote_share(body=body)
        print("The response of AgentApi->remote_share:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->remote_share: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RemoteShareRequest**](RemoteShareRequest.md)|  | [optional] 

### Return type

[**RemoteShare200Response**](RemoteShare200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remote_status**
> RemoteStatus200Response remote_status(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.enroll_request import EnrollRequest
from zrok_api.models.remote_status200_response import RemoteStatus200Response
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.EnrollRequest() # EnrollRequest |  (optional)

    try:
        api_response = api_instance.remote_status(body=body)
        print("The response of AgentApi->remote_status:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->remote_status: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnrollRequest**](EnrollRequest.md)|  | [optional] 

### Return type

[**RemoteStatus200Response**](RemoteStatus200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remote_unaccess**
> remote_unaccess(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.remote_unaccess_request import RemoteUnaccessRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.RemoteUnaccessRequest() # RemoteUnaccessRequest |  (optional)

    try:
        api_instance.remote_unaccess(body=body)
    except Exception as e:
        print("Exception when calling AgentApi->remote_unaccess: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RemoteUnaccessRequest**](RemoteUnaccessRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remote_unshare**
> remote_unshare(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.remote_unshare_request import RemoteUnshareRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.RemoteUnshareRequest() # RemoteUnshareRequest |  (optional)

    try:
        api_instance.remote_unshare(body=body)
    except Exception as e:
        print("Exception when calling AgentApi->remote_unshare: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RemoteUnshareRequest**](RemoteUnshareRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **share_http_healthcheck**
> ShareHttpHealthcheck200Response share_http_healthcheck(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.share_http_healthcheck200_response import ShareHttpHealthcheck200Response
from zrok_api.models.share_http_healthcheck_request import ShareHttpHealthcheckRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.ShareHttpHealthcheckRequest() # ShareHttpHealthcheckRequest |  (optional)

    try:
        api_response = api_instance.share_http_healthcheck(body=body)
        print("The response of AgentApi->share_http_healthcheck:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AgentApi->share_http_healthcheck: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ShareHttpHealthcheckRequest**](ShareHttpHealthcheckRequest.md)|  | [optional] 

### Return type

[**ShareHttpHealthcheck200Response**](ShareHttpHealthcheck200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |
**502** | bad gateway; agent not reachable |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unenroll**
> unenroll(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.enroll_request import EnrollRequest
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
    api_instance = zrok_api.AgentApi(api_client)
    body = zrok_api.EnrollRequest() # EnrollRequest |  (optional)

    try:
        api_instance.unenroll(body=body)
    except Exception as e:
        print("Exception when calling AgentApi->unenroll: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnrollRequest**](EnrollRequest.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**400** | bad request; not enrolled |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

