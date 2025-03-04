# zrok_api.EnvironmentApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**disable**](EnvironmentApi.md#disable) | **POST** /disable | 
[**enable**](EnvironmentApi.md#enable) | **POST** /enable | 

# **disable**
> disable(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# Configure API key authorization: key
configuration = zrok_api.Configuration()
configuration.api_key['x-token'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['x-token'] = 'Bearer'

# create an instance of the API class
api_instance = zrok_api.EnvironmentApi(zrok_api.ApiClient(configuration))
body = zrok_api.DisableBody() # DisableBody |  (optional)

try:
    api_instance.disable(body=body)
except ApiException as e:
    print("Exception when calling EnvironmentApi->disable: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DisableBody**](DisableBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **enable**
> InlineResponse2011 enable(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# Configure API key authorization: key
configuration = zrok_api.Configuration()
configuration.api_key['x-token'] = 'YOUR_API_KEY'
# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['x-token'] = 'Bearer'

# create an instance of the API class
api_instance = zrok_api.EnvironmentApi(zrok_api.ApiClient(configuration))
body = zrok_api.EnableBody() # EnableBody |  (optional)

try:
    api_response = api_instance.enable(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling EnvironmentApi->enable: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**EnableBody**](EnableBody.md)|  | [optional] 

### Return type

[**InlineResponse2011**](InlineResponse2011.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

