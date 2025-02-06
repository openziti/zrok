# zrok_api.ShareApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**access**](ShareApi.md#access) | **POST** /access | 
[**share**](ShareApi.md#share) | **POST** /share | 
[**unaccess**](ShareApi.md#unaccess) | **DELETE** /unaccess | 
[**unshare**](ShareApi.md#unshare) | **DELETE** /unshare | 
[**update_access**](ShareApi.md#update_access) | **PATCH** /access | 
[**update_share**](ShareApi.md#update_share) | **PATCH** /share | 

# **access**
> InlineResponse2013 access(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.AccessBody() # AccessBody |  (optional)

try:
    api_response = api_instance.access(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling ShareApi->access: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AccessBody**](AccessBody.md)|  | [optional] 

### Return type

[**InlineResponse2013**](InlineResponse2013.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **share**
> ShareResponse share(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.ShareRequest() # ShareRequest |  (optional)

try:
    api_response = api_instance.share(body=body)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unaccess**
> unaccess(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.UnaccessBody() # UnaccessBody |  (optional)

try:
    api_instance.unaccess(body=body)
except ApiException as e:
    print("Exception when calling ShareApi->unaccess: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnaccessBody**](UnaccessBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **unshare**
> unshare(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.UnshareBody() # UnshareBody |  (optional)

try:
    api_instance.unshare(body=body)
except ApiException as e:
    print("Exception when calling ShareApi->unshare: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UnshareBody**](UnshareBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_access**
> update_access(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.AccessBody1() # AccessBody1 |  (optional)

try:
    api_instance.update_access(body=body)
except ApiException as e:
    print("Exception when calling ShareApi->update_access: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AccessBody1**](AccessBody1.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_share**
> update_share(body=body)



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
api_instance = zrok_api.ShareApi(zrok_api.ApiClient(configuration))
body = zrok_api.ShareBody() # ShareBody |  (optional)

try:
    api_instance.update_share(body=body)
except ApiException as e:
    print("Exception when calling ShareApi->update_share: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ShareBody**](ShareBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

