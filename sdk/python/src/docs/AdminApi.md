# zrok_api.AdminApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**add_frontend_grant**](AdminApi.md#add_frontend_grant) | **POST** /frontend/grant | 
[**add_organization_member**](AdminApi.md#add_organization_member) | **POST** /organization/add | 
[**create_account**](AdminApi.md#create_account) | **POST** /account | 
[**create_frontend**](AdminApi.md#create_frontend) | **POST** /frontend | 
[**create_identity**](AdminApi.md#create_identity) | **POST** /identity | 
[**create_namespace**](AdminApi.md#create_namespace) | **POST** /namespace | 
[**create_organization**](AdminApi.md#create_organization) | **POST** /organization | 
[**delete_account**](AdminApi.md#delete_account) | **DELETE** /account | 
[**delete_frontend**](AdminApi.md#delete_frontend) | **DELETE** /frontend | 
[**delete_frontend_grant**](AdminApi.md#delete_frontend_grant) | **DELETE** /frontend/grant | 
[**delete_identity**](AdminApi.md#delete_identity) | **DELETE** /identity | 
[**delete_namespace**](AdminApi.md#delete_namespace) | **DELETE** /namespace | 
[**delete_organization**](AdminApi.md#delete_organization) | **DELETE** /organization | 
[**grants**](AdminApi.md#grants) | **POST** /grants | 
[**invite_token_generate**](AdminApi.md#invite_token_generate) | **POST** /invite/token/generate | 
[**list_frontends**](AdminApi.md#list_frontends) | **GET** /frontends | 
[**list_namespaces**](AdminApi.md#list_namespaces) | **GET** /namespaces | 
[**list_organization_members**](AdminApi.md#list_organization_members) | **POST** /organization/list | 
[**list_organizations**](AdminApi.md#list_organizations) | **GET** /organizations | 
[**remove_organization_member**](AdminApi.md#remove_organization_member) | **POST** /organization/remove | 
[**update_frontend**](AdminApi.md#update_frontend) | **PATCH** /frontend | 
[**update_namespace**](AdminApi.md#update_namespace) | **PATCH** /namespace | 


# **add_frontend_grant**
> add_frontend_grant(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.add_frontend_grant_request import AddFrontendGrantRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.AddFrontendGrantRequest() # AddFrontendGrantRequest |  (optional)

    try:
        api_instance.add_frontend_grant(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->add_frontend_grant: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AddFrontendGrantRequest**](AddFrontendGrantRequest.md)|  | [optional] 

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
**200** | ok |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **add_organization_member**
> add_organization_member(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.add_organization_member_request import AddOrganizationMemberRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.AddOrganizationMemberRequest() # AddOrganizationMemberRequest |  (optional)

    try:
        api_instance.add_organization_member(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->add_organization_member: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AddOrganizationMemberRequest**](AddOrganizationMemberRequest.md)|  | [optional] 

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
**201** | member added |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_account**
> RegenerateAccountToken200Response create_account(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.login_request import LoginRequest
from zrok_api.models.regenerate_account_token200_response import RegenerateAccountToken200Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.LoginRequest() # LoginRequest |  (optional)

    try:
        api_response = api_instance.create_account(body=body)
        print("The response of AdminApi->create_account:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->create_account: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**LoginRequest**](LoginRequest.md)|  | [optional] 

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
**201** | created |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_frontend**
> CreateFrontend201Response create_frontend(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_frontend201_response import CreateFrontend201Response
from zrok_api.models.create_frontend_request import CreateFrontendRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateFrontendRequest() # CreateFrontendRequest |  (optional)

    try:
        api_response = api_instance.create_frontend(body=body)
        print("The response of AdminApi->create_frontend:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->create_frontend: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateFrontendRequest**](CreateFrontendRequest.md)|  | [optional] 

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
**201** | frontend created |  -  |
**400** | bad request |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_identity**
> CreateIdentity201Response create_identity(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_identity201_response import CreateIdentity201Response
from zrok_api.models.create_identity_request import CreateIdentityRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateIdentityRequest() # CreateIdentityRequest |  (optional)

    try:
        api_response = api_instance.create_identity(body=body)
        print("The response of AdminApi->create_identity:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->create_identity: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateIdentityRequest**](CreateIdentityRequest.md)|  | [optional] 

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
**201** | created |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_namespace**
> CreateNamespace201Response create_namespace(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_namespace201_response import CreateNamespace201Response
from zrok_api.models.create_namespace_request import CreateNamespaceRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateNamespaceRequest() # CreateNamespaceRequest |  (optional)

    try:
        api_response = api_instance.create_namespace(body=body)
        print("The response of AdminApi->create_namespace:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->create_namespace: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateNamespaceRequest**](CreateNamespaceRequest.md)|  | [optional] 

### Return type

[**CreateNamespace201Response**](CreateNamespace201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | namespace created |  -  |
**401** | unauthorized |  -  |
**409** | namespace already exists |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **create_organization**
> CreateOrganization201Response create_organization(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_organization201_response import CreateOrganization201Response
from zrok_api.models.create_organization_request import CreateOrganizationRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateOrganizationRequest() # CreateOrganizationRequest |  (optional)

    try:
        api_response = api_instance.create_organization(body=body)
        print("The response of AdminApi->create_organization:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->create_organization: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateOrganizationRequest**](CreateOrganizationRequest.md)|  | [optional] 

### Return type

[**CreateOrganization201Response**](CreateOrganization201Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**201** | organization created |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_account**
> delete_account(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.verify200_response import Verify200Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.Verify200Response() # Verify200Response |  (optional)

    try:
        api_instance.delete_account(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_account: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Verify200Response**](Verify200Response.md)|  | [optional] 

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
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_frontend**
> delete_frontend(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_frontend201_response import CreateFrontend201Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateFrontend201Response() # CreateFrontend201Response |  (optional)

    try:
        api_instance.delete_frontend(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_frontend: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateFrontend201Response**](CreateFrontend201Response.md)|  | [optional] 

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
**200** | frontend deleted |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_frontend_grant**
> delete_frontend_grant(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.add_frontend_grant_request import AddFrontendGrantRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.AddFrontendGrantRequest() # AddFrontendGrantRequest |  (optional)

    try:
        api_instance.delete_frontend_grant(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_frontend_grant: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**AddFrontendGrantRequest**](AddFrontendGrantRequest.md)|  | [optional] 

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
**200** | ok |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_identity**
> delete_identity(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.delete_identity_request import DeleteIdentityRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.DeleteIdentityRequest() # DeleteIdentityRequest |  (optional)

    try:
        api_instance.delete_identity(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_identity: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**DeleteIdentityRequest**](DeleteIdentityRequest.md)|  | [optional] 

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
**200** | deleted |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_namespace**
> delete_namespace(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_namespace201_response import CreateNamespace201Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateNamespace201Response() # CreateNamespace201Response |  (optional)

    try:
        api_instance.delete_namespace(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_namespace: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateNamespace201Response**](CreateNamespace201Response.md)|  | [optional] 

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
**200** | namespace deleted |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **delete_organization**
> delete_organization(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_organization201_response import CreateOrganization201Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateOrganization201Response() # CreateOrganization201Response |  (optional)

    try:
        api_instance.delete_organization(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->delete_organization: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateOrganization201Response**](CreateOrganization201Response.md)|  | [optional] 

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
**200** | organization deleted |  -  |
**401** | unauthorized |  -  |
**404** | organization not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **grants**
> grants(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.verify200_response import Verify200Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.Verify200Response() # Verify200Response |  (optional)

    try:
        api_instance.grants(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->grants: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Verify200Response**](Verify200Response.md)|  | [optional] 

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
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **invite_token_generate**
> invite_token_generate(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.invite_token_generate_request import InviteTokenGenerateRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.InviteTokenGenerateRequest() # InviteTokenGenerateRequest |  (optional)

    try:
        api_instance.invite_token_generate(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->invite_token_generate: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**InviteTokenGenerateRequest**](InviteTokenGenerateRequest.md)|  | [optional] 

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
**201** | invite tokens created |  -  |
**400** | invite tokens not created |  -  |
**401** | unauthorized |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_frontends**
> List[ListFrontends200ResponseInner] list_frontends()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_frontends200_response_inner import ListFrontends200ResponseInner
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
    api_instance = zrok_api.AdminApi(api_client)

    try:
        api_response = api_instance.list_frontends()
        print("The response of AdminApi->list_frontends:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->list_frontends: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[ListFrontends200ResponseInner]**](ListFrontends200ResponseInner.md)

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

# **list_namespaces**
> List[ListNamespaces200ResponseInner] list_namespaces()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_namespaces200_response_inner import ListNamespaces200ResponseInner
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
    api_instance = zrok_api.AdminApi(api_client)

    try:
        api_response = api_instance.list_namespaces()
        print("The response of AdminApi->list_namespaces:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->list_namespaces: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**List[ListNamespaces200ResponseInner]**](ListNamespaces200ResponseInner.md)

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

# **list_organization_members**
> ListOrganizationMembers200Response list_organization_members(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.create_organization201_response import CreateOrganization201Response
from zrok_api.models.list_organization_members200_response import ListOrganizationMembers200Response
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.CreateOrganization201Response() # CreateOrganization201Response |  (optional)

    try:
        api_response = api_instance.list_organization_members(body=body)
        print("The response of AdminApi->list_organization_members:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->list_organization_members: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**CreateOrganization201Response**](CreateOrganization201Response.md)|  | [optional] 

### Return type

[**ListOrganizationMembers200Response**](ListOrganizationMembers200Response.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | list organization members |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **list_organizations**
> ListOrganizations200Response list_organizations()

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.list_organizations200_response import ListOrganizations200Response
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
    api_instance = zrok_api.AdminApi(api_client)

    try:
        api_response = api_instance.list_organizations()
        print("The response of AdminApi->list_organizations:\n")
        pprint(api_response)
    except Exception as e:
        print("Exception when calling AdminApi->list_organizations: %s\n" % e)
```



### Parameters

This endpoint does not need any parameter.

### Return type

[**ListOrganizations200Response**](ListOrganizations200Response.md)

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

# **remove_organization_member**
> remove_organization_member(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.remove_organization_member_request import RemoveOrganizationMemberRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.RemoveOrganizationMemberRequest() # RemoveOrganizationMemberRequest |  (optional)

    try:
        api_instance.remove_organization_member(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->remove_organization_member: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RemoveOrganizationMemberRequest**](RemoveOrganizationMemberRequest.md)|  | [optional] 

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
**200** | member removed |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_frontend**
> update_frontend(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.update_frontend_request import UpdateFrontendRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.UpdateFrontendRequest() # UpdateFrontendRequest |  (optional)

    try:
        api_instance.update_frontend(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->update_frontend: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateFrontendRequest**](UpdateFrontendRequest.md)|  | [optional] 

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
**200** | frontend updated |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **update_namespace**
> update_namespace(body=body)

### Example

* Api Key Authentication (key):

```python
import zrok_api
from zrok_api.models.update_namespace_request import UpdateNamespaceRequest
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
    api_instance = zrok_api.AdminApi(api_client)
    body = zrok_api.UpdateNamespaceRequest() # UpdateNamespaceRequest |  (optional)

    try:
        api_instance.update_namespace(body=body)
    except Exception as e:
        print("Exception when calling AdminApi->update_namespace: %s\n" % e)
```



### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**UpdateNamespaceRequest**](UpdateNamespaceRequest.md)|  | [optional] 

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
**200** | namespace updated |  -  |
**401** | unauthorized |  -  |
**404** | not found |  -  |
**500** | internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

