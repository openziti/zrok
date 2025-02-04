# zrok_api.AdminApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**add_organization_member**](AdminApi.md#add_organization_member) | **POST** /organization/add | 
[**create_account**](AdminApi.md#create_account) | **POST** /account | 
[**create_frontend**](AdminApi.md#create_frontend) | **POST** /frontend | 
[**create_identity**](AdminApi.md#create_identity) | **POST** /identity | 
[**create_organization**](AdminApi.md#create_organization) | **POST** /organization | 
[**delete_frontend**](AdminApi.md#delete_frontend) | **DELETE** /frontend | 
[**delete_organization**](AdminApi.md#delete_organization) | **DELETE** /organization | 
[**grants**](AdminApi.md#grants) | **POST** /grants | 
[**invite_token_generate**](AdminApi.md#invite_token_generate) | **POST** /invite/token/generate | 
[**list_frontends**](AdminApi.md#list_frontends) | **GET** /frontends | 
[**list_organization_members**](AdminApi.md#list_organization_members) | **POST** /organization/list | 
[**list_organizations**](AdminApi.md#list_organizations) | **GET** /organizations | 
[**remove_organization_member**](AdminApi.md#remove_organization_member) | **POST** /organization/remove | 
[**update_frontend**](AdminApi.md#update_frontend) | **PATCH** /frontend | 

# **add_organization_member**
> add_organization_member(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.OrganizationAddBody() # OrganizationAddBody |  (optional)

try:
    api_instance.add_organization_member(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->add_organization_member: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**OrganizationAddBody**](OrganizationAddBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_account**
> InlineResponse200 create_account(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.AccountBody() # AccountBody |  (optional)

try:
    api_response = api_instance.create_account(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->create_account: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AccountBody**](AccountBody.md)|  | [optional] 

### Return type

[**InlineResponse200**](InlineResponse200.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_frontend**
> InlineResponse201 create_frontend(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.FrontendBody() # FrontendBody |  (optional)

try:
    api_response = api_instance.create_frontend(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->create_frontend: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**FrontendBody**](FrontendBody.md)|  | [optional] 

### Return type

[**InlineResponse201**](InlineResponse201.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_identity**
> InlineResponse2011 create_identity(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.IdentityBody() # IdentityBody |  (optional)

try:
    api_response = api_instance.create_identity(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->create_identity: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**IdentityBody**](IdentityBody.md)|  | [optional] 

### Return type

[**InlineResponse2011**](InlineResponse2011.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_organization**
> InlineResponse2012 create_organization(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.OrganizationBody() # OrganizationBody |  (optional)

try:
    api_response = api_instance.create_organization(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->create_organization: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**OrganizationBody**](OrganizationBody.md)|  | [optional] 

### Return type

[**InlineResponse2012**](InlineResponse2012.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_frontend**
> delete_frontend(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.FrontendBody1() # FrontendBody1 |  (optional)

try:
    api_instance.delete_frontend(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->delete_frontend: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**FrontendBody1**](FrontendBody1.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_organization**
> delete_organization(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.OrganizationBody1() # OrganizationBody1 |  (optional)

try:
    api_instance.delete_organization(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->delete_organization: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**OrganizationBody1**](OrganizationBody1.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **grants**
> grants(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.GrantsBody() # GrantsBody |  (optional)

try:
    api_instance.grants(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->grants: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**GrantsBody**](GrantsBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **invite_token_generate**
> invite_token_generate(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.TokenGenerateBody() # TokenGenerateBody |  (optional)

try:
    api_instance.invite_token_generate(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->invite_token_generate: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**TokenGenerateBody**](TokenGenerateBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_frontends**
> list[InlineResponse2002] list_frontends()



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))

try:
    api_response = api_instance.list_frontends()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->list_frontends: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**list[InlineResponse2002]**](InlineResponse2002.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_organization_members**
> InlineResponse2003 list_organization_members(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.OrganizationListBody() # OrganizationListBody |  (optional)

try:
    api_response = api_instance.list_organization_members(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->list_organization_members: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**OrganizationListBody**](OrganizationListBody.md)|  | [optional] 

### Return type

[**InlineResponse2003**](InlineResponse2003.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_organizations**
> InlineResponse2004 list_organizations()



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))

try:
    api_response = api_instance.list_organizations()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AdminApi->list_organizations: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**InlineResponse2004**](InlineResponse2004.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **remove_organization_member**
> remove_organization_member(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.OrganizationRemoveBody() # OrganizationRemoveBody |  (optional)

try:
    api_instance.remove_organization_member(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->remove_organization_member: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**OrganizationRemoveBody**](OrganizationRemoveBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_frontend**
> update_frontend(body=body)



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
api_instance = zrok_api.AdminApi(zrok_api.ApiClient(configuration))
body = zrok_api.FrontendBody2() # FrontendBody2 |  (optional)

try:
    api_instance.update_frontend(body=body)
except ApiException as e:
    print("Exception when calling AdminApi->update_frontend: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**FrontendBody2**](FrontendBody2.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

