# zrok_api.AgentApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ping**](AgentApi.md#ping) | **POST** /agent/ping | 


# **ping**
> Ping200Response ping(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.ping200_response import Ping200Response
from zrok_api.models.ping_request import PingRequest
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
    body = zrok_api.PingRequest() # PingRequest |  (optional)

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
 **body** | [**PingRequest**](PingRequest.md)|  | [optional] 

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

