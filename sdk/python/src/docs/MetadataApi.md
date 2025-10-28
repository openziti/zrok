# zrok_api.MetadataApi

All URIs are relative to */api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**client_version_check**](MetadataApi.md#client_version_check) | **POST** /clientVersionCheck | 
[**configuration**](MetadataApi.md#configuration) | **GET** /configuration | 
[**get_account_detail**](MetadataApi.md#get_account_detail) | **GET** /detail/account | 
[**get_account_metrics**](MetadataApi.md#get_account_metrics) | **GET** /metrics/account | 
[**get_environment_detail**](MetadataApi.md#get_environment_detail) | **GET** /detail/environment/{envZId} | 
[**get_environment_metrics**](MetadataApi.md#get_environment_metrics) | **GET** /metrics/environment/{envId} | 
[**get_frontend_detail**](MetadataApi.md#get_frontend_detail) | **GET** /detail/frontend/{frontendId} | 
[**get_share_detail**](MetadataApi.md#get_share_detail) | **GET** /detail/share/{shareToken} | 
[**get_share_metrics**](MetadataApi.md#get_share_metrics) | **GET** /metrics/share/{shareToken} | 
[**get_sparklines**](MetadataApi.md#get_sparklines) | **POST** /sparklines | 
[**list_accesses**](MetadataApi.md#list_accesses) | **GET** /accesses | 
[**list_environments**](MetadataApi.md#list_environments) | **GET** /environments | 
[**list_memberships**](MetadataApi.md#list_memberships) | **GET** /memberships | 
[**list_org_members**](MetadataApi.md#list_org_members) | **GET** /members/{organizationToken} | 
[**list_shares**](MetadataApi.md#list_shares) | **GET** /shares | 
[**org_account_overview**](MetadataApi.md#org_account_overview) | **GET** /overview/{organizationToken}/{accountEmail} | 
[**overview**](MetadataApi.md#overview) | **GET** /overview | 
[**version**](MetadataApi.md#version) | **GET** /version | 
[**version_inventory**](MetadataApi.md#version_inventory) | **GET** /versions | 


# **client_version_check**
> client_version_check(body=body)

### Example


```python
import zrok_api
from zrok_api.models.client_version_check_request import ClientVersionCheckRequest
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.MetadataApi(api_client)
    body = zrok_api.ClientVersionCheckRequest() # ClientVersionCheckRequest |  (optional)

    try:
        api_instance.client_version_check(body=body)
    except Exception as e:
        print("Exception when calling MetadataApi->client_version_check: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ClientVersionCheckRequest**](ClientVersionCheckRequest.md)|  | [optional] 

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
**200** | compatible |  -  |
**400** | not compatible |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **configuration**
> Configuration configuration()

### Example


```python
import zrok_api
from zrok_api.models.configuration import Configuration
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.configuration()
        print("The response of MetadataApi->configuration:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->configuration: %s\n" % e)
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

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | current configuration |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_account_detail**
> List[Environment] get_account_detail()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.environment import Environment
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
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.get_account_detail()
        print("The response of MetadataApi->get_account_detail:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_account_detail: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[Environment]**](Environment.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_account_metrics**
> Metrics get_account_metrics(duration=duration)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.metrics import Metrics
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
    api_instance = zrok_api.MetadataApi(api_client)
    duration = 'duration_example' # str |  (optional)

    try:
        api_response = api_instance.get_account_metrics(duration=duration)
        print("The response of MetadataApi->get_account_metrics:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_account_metrics: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **duration** | **str**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | account metrics |  -  |
**400** | bad request |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_environment_detail**
> EnvironmentAndResources get_environment_detail(env_zid)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.environment_and_resources import EnvironmentAndResources
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
    api_instance = zrok_api.MetadataApi(api_client)
    env_zid = 'env_zid_example' # str | 

    try:
        api_response = api_instance.get_environment_detail(env_zid)
        print("The response of MetadataApi->get_environment_detail:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_environment_detail: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env_zid** | **str**|  | 

### Return type

[**EnvironmentAndResources**](EnvironmentAndResources.md)

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
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_environment_metrics**
> Metrics get_environment_metrics(env_id, duration=duration)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.metrics import Metrics
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
    api_instance = zrok_api.MetadataApi(api_client)
    env_id = 'env_id_example' # str | 
    duration = 'duration_example' # str |  (optional)

    try:
        api_response = api_instance.get_environment_metrics(env_id, duration=duration)
        print("The response of MetadataApi->get_environment_metrics:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_environment_metrics: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env_id** | **str**|  | 
 **duration** | **str**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | environment metrics |  -  |
**400** | bad request |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_frontend_detail**
> Frontend get_frontend_detail(frontend_id)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.frontend import Frontend
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
    api_instance = zrok_api.MetadataApi(api_client)
    frontend_id = 56 # int | 

    try:
        api_response = api_instance.get_frontend_detail(frontend_id)
        print("The response of MetadataApi->get_frontend_detail:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_frontend_detail: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **frontend_id** | **int**|  | 

### Return type

[**Frontend**](Frontend.md)

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
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_share_detail**
> Share get_share_detail(share_token)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.share import Share
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
    api_instance = zrok_api.MetadataApi(api_client)
    share_token = 'share_token_example' # str | 

    try:
        api_response = api_instance.get_share_detail(share_token)
        print("The response of MetadataApi->get_share_detail:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_share_detail: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **share_token** | **str**|  | 

### Return type

[**Share**](Share.md)

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
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_share_metrics**
> Metrics get_share_metrics(share_token, duration=duration)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.metrics import Metrics
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
    api_instance = zrok_api.MetadataApi(api_client)
    share_token = 'share_token_example' # str | 
    duration = 'duration_example' # str |  (optional)

    try:
        api_response = api_instance.get_share_metrics(share_token, duration=duration)
        print("The response of MetadataApi->get_share_metrics:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_share_metrics: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **share_token** | **str**|  | 
 **duration** | **str**|  | [optional] 

### Return type

[**Metrics**](Metrics.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | share metrics |  -  |
**400** | bad request |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_sparklines**
> GetSparklines200Response get_sparklines(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.get_sparklines200_response import GetSparklines200Response
from zrok_api.models.get_sparklines_request import GetSparklinesRequest
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
    api_instance = zrok_api.MetadataApi(api_client)
    body = zrok_api.GetSparklinesRequest() # GetSparklinesRequest |  (optional)

    try:
        api_response = api_instance.get_sparklines(body=body)
        print("The response of MetadataApi->get_sparklines:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->get_sparklines: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**GetSparklinesRequest**](GetSparklinesRequest.md)|  | [optional] 

### Return type

[**GetSparklines200Response**](GetSparklines200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | sparklines data |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_accesses**
> AccessesList list_accesses(env_zid=env_zid, share_token=share_token, bind_address=bind_address, description=description, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.accesses_list import AccessesList
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
    api_instance = zrok_api.MetadataApi(api_client)
    env_zid = 'env_zid_example' # str | filter by environment ziti identity (optional)
    share_token = 'share_token_example' # str | filter by associated share token (optional)
    bind_address = 'bind_address_example' # str | filter by bind address (substring match) (optional)
    description = 'description_example' # str | filter by description (substring match) (optional)
    created_after = 'created_after_example' # str | filter by created date (RFC3339 format) (optional)
    created_before = 'created_before_example' # str | filter by created date (RFC3339 format) (optional)
    updated_after = 'updated_after_example' # str | filter by updated date (RFC3339 format) (optional)
    updated_before = 'updated_before_example' # str | filter by updated date (RFC3339 format) (optional)

    try:
        api_response = api_instance.list_accesses(env_zid=env_zid, share_token=share_token, bind_address=bind_address, description=description, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before)
        print("The response of MetadataApi->list_accesses:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->list_accesses: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env_zid** | **str**| filter by environment ziti identity | [optional] 
 **share_token** | **str**| filter by associated share token | [optional] 
 **bind_address** | **str**| filter by bind address (substring match) | [optional] 
 **description** | **str**| filter by description (substring match) | [optional] 
 **created_after** | **str**| filter by created date (RFC3339 format) | [optional] 
 **created_before** | **str**| filter by created date (RFC3339 format) | [optional] 
 **updated_after** | **str**| filter by updated date (RFC3339 format) | [optional] 
 **updated_before** | **str**| filter by updated date (RFC3339 format) | [optional] 

### Return type

[**AccessesList**](AccessesList.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | list of accesses |  -  |
**400** | bad request (e.g., activityDuration exceeds 30d, invalid date format) |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_environments**
> EnvironmentsList list_environments(description=description, host=host, address=address, remote_agent=remote_agent, has_shares=has_shares, has_accesses=has_accesses, has_activity=has_activity, share_count=share_count, access_count=access_count, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before, activity_duration=activity_duration)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.environments_list import EnvironmentsList
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
    api_instance = zrok_api.MetadataApi(api_client)
    description = 'description_example' # str | filter by description (case-insensitive substring match) (optional)
    host = 'host_example' # str | filter by host (case-insensitive substring match) (optional)
    address = 'address_example' # str | filter by address (exact match) (optional)
    remote_agent = True # bool | filter by whether agent is enrolled (optional)
    has_shares = True # bool | filter by whether environment has active shares (optional)
    has_accesses = True # bool | filter by whether environment has active accesses (optional)
    has_activity = True # bool | filter by whether environment has metrics within activityDuration timeframe (optional)
    share_count = 'share_count_example' # str | filter by share count with operator (e.g., \">0\", \">=5\", \"=0\", \"<10\", \"<=3\") (optional)
    access_count = 'access_count_example' # str | filter by access count with operator (e.g., \">0\", \">=5\", \"=0\", \"<10\", \"<=3\") (optional)
    created_after = 'created_after_example' # str | filter by created date (RFC3339 datetime, inclusive) (optional)
    created_before = 'created_before_example' # str | filter by created date (RFC3339 datetime, inclusive) (optional)
    updated_after = 'updated_after_example' # str | filter by updated date (RFC3339 datetime, inclusive) (optional)
    updated_before = 'updated_before_example' # str | filter by updated date (RFC3339 datetime, inclusive) (optional)
    activity_duration = 'activity_duration_example' # str | duration for hasActivity filter (e.g., \"24h\", \"7d\", \"30d\"). default \"24h\", maximum \"30d\" (720h) (optional)

    try:
        api_response = api_instance.list_environments(description=description, host=host, address=address, remote_agent=remote_agent, has_shares=has_shares, has_accesses=has_accesses, has_activity=has_activity, share_count=share_count, access_count=access_count, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before, activity_duration=activity_duration)
        print("The response of MetadataApi->list_environments:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->list_environments: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **description** | **str**| filter by description (case-insensitive substring match) | [optional] 
 **host** | **str**| filter by host (case-insensitive substring match) | [optional] 
 **address** | **str**| filter by address (exact match) | [optional] 
 **remote_agent** | **bool**| filter by whether agent is enrolled | [optional] 
 **has_shares** | **bool**| filter by whether environment has active shares | [optional] 
 **has_accesses** | **bool**| filter by whether environment has active accesses | [optional] 
 **has_activity** | **bool**| filter by whether environment has metrics within activityDuration timeframe | [optional] 
 **share_count** | **str**| filter by share count with operator (e.g., \&quot;&gt;0\&quot;, \&quot;&gt;&#x3D;5\&quot;, \&quot;&#x3D;0\&quot;, \&quot;&lt;10\&quot;, \&quot;&lt;&#x3D;3\&quot;) | [optional] 
 **access_count** | **str**| filter by access count with operator (e.g., \&quot;&gt;0\&quot;, \&quot;&gt;&#x3D;5\&quot;, \&quot;&#x3D;0\&quot;, \&quot;&lt;10\&quot;, \&quot;&lt;&#x3D;3\&quot;) | [optional] 
 **created_after** | **str**| filter by created date (RFC3339 datetime, inclusive) | [optional] 
 **created_before** | **str**| filter by created date (RFC3339 datetime, inclusive) | [optional] 
 **updated_after** | **str**| filter by updated date (RFC3339 datetime, inclusive) | [optional] 
 **updated_before** | **str**| filter by updated date (RFC3339 datetime, inclusive) | [optional] 
 **activity_duration** | **str**| duration for hasActivity filter (e.g., \&quot;24h\&quot;, \&quot;7d\&quot;, \&quot;30d\&quot;). default \&quot;24h\&quot;, maximum \&quot;30d\&quot; (720h) | [optional] 

### Return type

[**EnvironmentsList**](EnvironmentsList.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | list of environments |  -  |
**400** | bad request (e.g., activityDuration exceeds 30d, invalid date format, invalid operator) |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_memberships**
> ListMemberships200Response list_memberships()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_memberships200_response import ListMemberships200Response
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
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.list_memberships()
        print("The response of MetadataApi->list_memberships:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->list_memberships: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**ListMemberships200Response**](ListMemberships200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_org_members**
> ListOrganizationMembers200Response list_org_members(organization_token)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_organization_members200_response import ListOrganizationMembers200Response
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
    api_instance = zrok_api.MetadataApi(api_client)
    organization_token = 'organization_token_example' # str | 

    try:
        api_response = api_instance.list_org_members(organization_token)
        print("The response of MetadataApi->list_org_members:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->list_org_members: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **organization_token** | **str**|  | 

### Return type

[**ListOrganizationMembers200Response**](ListOrganizationMembers200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_shares**
> SharesList list_shares(env_zid=env_zid, share_mode=share_mode, backend_mode=backend_mode, share_token=share_token, target=target, permission_mode=permission_mode, has_activity=has_activity, activity_duration=activity_duration, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.shares_list import SharesList
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
    api_instance = zrok_api.MetadataApi(api_client)
    env_zid = 'env_zid_example' # str | filter by environment ziti identity (optional)
    share_mode = 'share_mode_example' # str | filter by share mode (public/private) (optional)
    backend_mode = 'backend_mode_example' # str | filter by backend mode (optional)
    share_token = 'share_token_example' # str | filter by share token (substring match) (optional)
    target = 'target_example' # str | filter by target (substring match) (optional)
    permission_mode = 'permission_mode_example' # str | filter by permission mode (open/closed) (optional)
    has_activity = True # bool | filter shares with recent activity (optional)
    activity_duration = 'activity_duration_example' # str | duration for hasActivity filter (e.g., \"24h\", \"7d\", \"30d\"). default \"24h\", maximum \"30d\" (720h) (optional)
    created_after = 'created_after_example' # str | filter by created date (RFC3339 format) (optional)
    created_before = 'created_before_example' # str | filter by created date (RFC3339 format) (optional)
    updated_after = 'updated_after_example' # str | filter by updated date (RFC3339 format) (optional)
    updated_before = 'updated_before_example' # str | filter by updated date (RFC3339 format) (optional)

    try:
        api_response = api_instance.list_shares(env_zid=env_zid, share_mode=share_mode, backend_mode=backend_mode, share_token=share_token, target=target, permission_mode=permission_mode, has_activity=has_activity, activity_duration=activity_duration, created_after=created_after, created_before=created_before, updated_after=updated_after, updated_before=updated_before)
        print("The response of MetadataApi->list_shares:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->list_shares: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **env_zid** | **str**| filter by environment ziti identity | [optional] 
 **share_mode** | **str**| filter by share mode (public/private) | [optional] 
 **backend_mode** | **str**| filter by backend mode | [optional] 
 **share_token** | **str**| filter by share token (substring match) | [optional] 
 **target** | **str**| filter by target (substring match) | [optional] 
 **permission_mode** | **str**| filter by permission mode (open/closed) | [optional] 
 **has_activity** | **bool**| filter shares with recent activity | [optional] 
 **activity_duration** | **str**| duration for hasActivity filter (e.g., \&quot;24h\&quot;, \&quot;7d\&quot;, \&quot;30d\&quot;). default \&quot;24h\&quot;, maximum \&quot;30d\&quot; (720h) | [optional] 
 **created_after** | **str**| filter by created date (RFC3339 format) | [optional] 
 **created_before** | **str**| filter by created date (RFC3339 format) | [optional] 
 **updated_after** | **str**| filter by updated date (RFC3339 format) | [optional] 
 **updated_before** | **str**| filter by updated date (RFC3339 format) | [optional] 

### Return type

[**SharesList**](SharesList.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | list of shares |  -  |
**400** | bad request (e.g., activityDuration exceeds 30d, invalid date format) |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **org_account_overview**
> Overview org_account_overview(organization_token, account_email)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.overview import Overview
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
    api_instance = zrok_api.MetadataApi(api_client)
    organization_token = 'organization_token_example' # str | 
    account_email = 'account_email_example' # str | 

    try:
        api_response = api_instance.org_account_overview(organization_token, account_email)
        print("The response of MetadataApi->org_account_overview:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->org_account_overview: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **organization_token** | **str**|  | 
 **account_email** | **str**|  | 

### Return type

[**Overview**](Overview.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **overview**
> Overview overview()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.overview import Overview
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
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.overview()
        print("The response of MetadataApi->overview:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->overview: %s\n" % e)
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

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | overview returned |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **version**
> str version()

### Example


```python
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.version()
        print("The response of MetadataApi->version:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->version: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | legacy upgrade required |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **version_inventory**
> VersionInventory200Response version_inventory()

### Example


```python
import zrok_api
from zrok_api.models.version_inventory200_response import VersionInventory200Response
from zrok_api.rest import ApiException
from pprint import pprint

# Defining the host is optional and defaults to /api/v2
# See configuration.py for a list of all supported configuration parameters.
configuration = zrok_api.Configuration(
    host = "/api/v2"
)


# Enter a context with an instance of the API client
with zrok_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = zrok_api.MetadataApi(api_client)

    try:
        api_response = api_instance.version_inventory()
        print("The response of MetadataApi->version_inventory:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling MetadataApi->version_inventory: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**VersionInventory200Response**](VersionInventory200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | ok |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

