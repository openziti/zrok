# zrok_api.ShareApi

All URIs are relative to */api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**access**](ShareApi.md#access) | **POST** /access | 
[**create_share_name**](ShareApi.md#create_share_name) | **POST** /share/name | 
[**delete_share_name**](ShareApi.md#delete_share_name) | **DELETE** /share/name | 
[**list_all_names**](ShareApi.md#list_all_names) | **GET** /share/names | 
[**list_names_for_namespace**](ShareApi.md#list_names_for_namespace) | **GET** /share/names/{namespaceToken} | 
[**list_share_namespaces**](ShareApi.md#list_share_namespaces) | **GET** /share/namespaces | 
[**share**](ShareApi.md#share) | **POST** /share | 
[**unaccess**](ShareApi.md#unaccess) | **DELETE** /unaccess | 
[**unshare**](ShareApi.md#unshare) | **DELETE** /unshare | 
[**update_access**](ShareApi.md#update_access) | **PATCH** /access | 
[**update_share**](ShareApi.md#update_share) | **PATCH** /share | 
[**update_share_name**](ShareApi.md#update_share_name) | **PATCH** /share/name | 


# **access**
> Access201Response access(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.access201_response import Access201Response
from zrok_api.models.access_request import AccessRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.AccessRequest() # AccessRequest |  (optional)

    try:
        api_response = api_instance.access(body=body)
        print("The response of ShareApi->access:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ShareApi->access: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AccessRequest**](AccessRequest.md)|  | [optional] 

### Return type

[**Access201Response**](Access201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | access created |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_share_name**
> create_share_name(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_share_name_request import CreateShareNameRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.CreateShareNameRequest() # CreateShareNameRequest |  (optional)

    try:
        api_instance.create_share_name(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->create_share_name: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateShareNameRequest**](CreateShareNameRequest.md)|  | [optional] 

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
**201** | name created |  -  |
**401** | unauthorized |  -  |
**404** | namespace not found |  -  |
**409** | name already exists |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_share_name**
> delete_share_name(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_share_name_request import CreateShareNameRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.CreateShareNameRequest() # CreateShareNameRequest |  (optional)

    try:
        api_instance.delete_share_name(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->delete_share_name: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateShareNameRequest**](CreateShareNameRequest.md)|  | [optional] 

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
**200** | name deleted |  -  |
**401** | unauthorized |  -  |
**404** | name not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_all_names**
> List[Name] list_all_names()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.name import Name
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)

    try:
        api_response = api_instance.list_all_names()
        print("The response of ShareApi->list_all_names:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ShareApi->list_all_names: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[Name]**](Name.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_names_for_namespace**
> List[Name] list_names_for_namespace(namespace_token)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.name import Name
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    namespace_token = 'namespace_token_example' # str | 

    try:
        api_response = api_instance.list_names_for_namespace(namespace_token)
        print("The response of ShareApi->list_names_for_namespace:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ShareApi->list_names_for_namespace: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **namespace_token** | **str**|  | 

### Return type

[**List[Name]**](Name.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**404** | namespace not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_share_namespaces**
> List[ListShareNamespaces200ResponseInner] list_share_namespaces()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_share_namespaces200_response_inner import ListShareNamespaces200ResponseInner
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)

    try:
        api_response = api_instance.list_share_namespaces()
        print("The response of ShareApi->list_share_namespaces:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ShareApi->list_share_namespaces: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[ListShareNamespaces200ResponseInner]**](ListShareNamespaces200ResponseInner.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **share**
> ShareResponse share(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.share_request import ShareRequest
from zrok_api.models.share_response import ShareResponse
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.ShareRequest() # ShareRequest |  (optional)

    try:
        api_response = api_instance.share(body=body)
        print("The response of ShareApi->share:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling ShareApi->share: %s\n" % e)
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

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | share created |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**409** | conflict |  -  |
**422** | unprocessable |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unaccess**
> unaccess(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.unaccess_request import UnaccessRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.UnaccessRequest() # UnaccessRequest |  (optional)

    try:
        api_instance.unaccess(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->unaccess: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnaccessRequest**](UnaccessRequest.md)|  | [optional] 

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
**200** | access removed |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unshare**
> unshare(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.unshare_request import UnshareRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.UnshareRequest() # UnshareRequest |  (optional)

    try:
        api_instance.unshare(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->unshare: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnshareRequest**](UnshareRequest.md)|  | [optional] 

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
**200** | share removed |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_access**
> update_access(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.update_access_request import UpdateAccessRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.UpdateAccessRequest() # UpdateAccessRequest |  (optional)

    try:
        api_instance.update_access(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->update_access: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateAccessRequest**](UpdateAccessRequest.md)|  | [optional] 

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
**200** | access updated |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_share**
> update_share(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.update_share_request import UpdateShareRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.UpdateShareRequest() # UpdateShareRequest |  (optional)

    try:
        api_instance.update_share(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->update_share: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateShareRequest**](UpdateShareRequest.md)|  | [optional] 

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
**200** | share updated |  -  |
**400** | bad request |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_share_name**
> update_share_name(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.update_share_name_request import UpdateShareNameRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
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
    api_instance = zrok_api.ShareApi(api_client)
    body = zrok_api.UpdateShareNameRequest() # UpdateShareNameRequest |  (optional)

    try:
        api_instance.update_share_name(body=body)
    except Exception as e:
        print("Exception when calling ShareApi->update_share_name: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateShareNameRequest**](UpdateShareNameRequest.md)|  | [optional] 

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
**200** | name updated successfully |  -  |
**401** | unauthorized |  -  |
**404** | name not found |  -  |
**409** | conflict - cannot change reservation state |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

