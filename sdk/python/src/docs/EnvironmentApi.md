# zrok_api.EnvironmentApi

All URIs are relative to */api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**disable**](EnvironmentApi.md#disable) | **POST** /disable | 
[**enable**](EnvironmentApi.md#enable) | **POST** /enable | 


# **disable**
> disable(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.disable_request import DisableRequest
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
    api_instance = zrok_api.EnvironmentApi(api_client)
    body = zrok_api.DisableRequest() # DisableRequest |  (optional)

    try:
        api_instance.disable(body=body)
    except Exception as e:
        print("Exception when calling EnvironmentApi->disable: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DisableRequest**](DisableRequest.md)|  | [optional] 

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
**200** | environment disabled |  -  |
**401** | invalid environment |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **enable**
> CreateIdentity201Response enable(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_identity201_response import CreateIdentity201Response
from zrok_api.models.enable_request import EnableRequest
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
    api_instance = zrok_api.EnvironmentApi(api_client)
    body = zrok_api.EnableRequest() # EnableRequest |  (optional)

    try:
        api_response = api_instance.enable(body=body)
        print("The response of EnvironmentApi->enable:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling EnvironmentApi->enable: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnableRequest**](EnableRequest.md)|  | [optional] 

### Return type

[**CreateIdentity201Response**](CreateIdentity201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | environment enabled |  -  |
**401** | unauthorized |  -  |
**404** | account not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

