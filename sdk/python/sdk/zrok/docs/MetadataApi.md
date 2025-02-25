# zrok_api.MetadataApi

All URIs are relative to */api/v1*

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
[**list_memberships**](MetadataApi.md#list_memberships) | **GET** /memberships | 
[**list_org_members**](MetadataApi.md#list_org_members) | **GET** /members/{organizationToken} | 
[**org_account_overview**](MetadataApi.md#org_account_overview) | **GET** /overview/{organizationToken}/{accountEmail} | 
[**overview**](MetadataApi.md#overview) | **GET** /overview | 
[**version**](MetadataApi.md#version) | **GET** /version | 
[**version_inventory**](MetadataApi.md#version_inventory) | **GET** /versions | 

# **client_version_check**
> client_version_check(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.MetadataApi()
body = zrok_api.ClientVersionCheckBody() # ClientVersionCheckBody |  (optional)

try:
    api_instance.client_version_check(body=body)
except ApiException as e:
    print("Exception when calling MetadataApi->client_version_check: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ClientVersionCheckBody**](ClientVersionCheckBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **configuration**
> Configuration configuration()



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.MetadataApi()

try:
    api_response = api_instance.configuration()
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_account_detail**
> Environments get_account_detail()



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))

try:
    api_response = api_instance.get_account_detail()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->get_account_detail: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Environments**](Environments.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_account_metrics**
> Metrics get_account_metrics(duration=duration)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
duration = 'duration_example' # str |  (optional)

try:
    api_response = api_instance.get_account_metrics(duration=duration)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_environment_detail**
> EnvironmentAndResources get_environment_detail(env_zid)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
env_zid = 'env_zid_example' # str | 

try:
    api_response = api_instance.get_environment_detail(env_zid)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_environment_metrics**
> Metrics get_environment_metrics(env_id, duration=duration)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
env_id = 'env_id_example' # str | 
duration = 'duration_example' # str |  (optional)

try:
    api_response = api_instance.get_environment_metrics(env_id, duration=duration)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_frontend_detail**
> Frontend get_frontend_detail(frontend_id)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
frontend_id = 56 # int | 

try:
    api_response = api_instance.get_frontend_detail(frontend_id)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_share_detail**
> Share get_share_detail(share_token)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
share_token = 'share_token_example' # str | 

try:
    api_response = api_instance.get_share_detail(share_token)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_share_metrics**
> Metrics get_share_metrics(share_token, duration=duration)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
share_token = 'share_token_example' # str | 
duration = 'duration_example' # str |  (optional)

try:
    api_response = api_instance.get_share_metrics(share_token, duration=duration)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_sparklines**
> InlineResponse2006 get_sparklines(body=body)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
body = zrok_api.SparklinesBody() # SparklinesBody |  (optional)

try:
    api_response = api_instance.get_sparklines(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->get_sparklines: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**SparklinesBody**](SparklinesBody.md)|  | [optional] 

### Return type

[**InlineResponse2006**](InlineResponse2006.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_memberships**
> InlineResponse2005 list_memberships()



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))

try:
    api_response = api_instance.list_memberships()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->list_memberships: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**InlineResponse2005**](InlineResponse2005.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_org_members**
> InlineResponse2003 list_org_members(organization_token)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
organization_token = 'organization_token_example' # str | 

try:
    api_response = api_instance.list_org_members(organization_token)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->list_org_members: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **organization_token** | **str**|  | 

### Return type

[**InlineResponse2003**](InlineResponse2003.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **org_account_overview**
> Overview org_account_overview(organization_token, account_email)



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))
organization_token = 'organization_token_example' # str | 
account_email = 'account_email_example' # str | 

try:
    api_response = api_instance.org_account_overview(organization_token, account_email)
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **overview**
> Overview overview()



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
api_instance = zrok_api.MetadataApi(zrok_api.ApiClient(configuration))

try:
    api_response = api_instance.overview()
    pprint(api_response)
except ApiException as e:
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

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **version**
> Version version()



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.MetadataApi()

try:
    api_response = api_instance.version()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->version: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Version**](Version.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **version_inventory**
> InlineResponse2007 version_inventory()



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.MetadataApi()

try:
    api_response = api_instance.version_inventory()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling MetadataApi->version_inventory: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**InlineResponse2007**](InlineResponse2007.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

