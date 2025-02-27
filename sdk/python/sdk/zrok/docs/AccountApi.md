# zrok_api.AccountApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**change_password**](AccountApi.md#change_password) | **POST** /changePassword | 
[**invite**](AccountApi.md#invite) | **POST** /invite | 
[**login**](AccountApi.md#login) | **POST** /login | 
[**regenerate_account_token**](AccountApi.md#regenerate_account_token) | **POST** /regenerateAccountToken | 
[**register**](AccountApi.md#register) | **POST** /register | 
[**reset_password**](AccountApi.md#reset_password) | **POST** /resetPassword | 
[**reset_password_request**](AccountApi.md#reset_password_request) | **POST** /resetPasswordRequest | 
[**verify**](AccountApi.md#verify) | **POST** /verify | 

# **change_password**
> change_password(body=body)



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
api_instance = zrok_api.AccountApi(zrok_api.ApiClient(configuration))
body = zrok_api.ChangePasswordBody() # ChangePasswordBody |  (optional)

try:
    api_instance.change_password(body=body)
except ApiException as e:
    print("Exception when calling AccountApi->change_password: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ChangePasswordBody**](ChangePasswordBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **invite**
> invite(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.InviteBody() # InviteBody |  (optional)

try:
    api_instance.invite(body=body)
except ApiException as e:
    print("Exception when calling AccountApi->invite: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**InviteBody**](InviteBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> str login(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.LoginBody() # LoginBody |  (optional)

try:
    api_response = api_instance.login(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AccountApi->login: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**LoginBody**](LoginBody.md)|  | [optional] 

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **regenerate_account_token**
> InlineResponse200 regenerate_account_token(body=body)



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
api_instance = zrok_api.AccountApi(zrok_api.ApiClient(configuration))
body = zrok_api.RegenerateAccountTokenBody() # RegenerateAccountTokenBody |  (optional)

try:
    api_response = api_instance.regenerate_account_token(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AccountApi->regenerate_account_token: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegenerateAccountTokenBody**](RegenerateAccountTokenBody.md)|  | [optional] 

### Return type

[**InlineResponse200**](InlineResponse200.md)

### Authorization

[key](../README.md#key)

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **register**
> InlineResponse200 register(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.RegisterBody() # RegisterBody |  (optional)

try:
    api_response = api_instance.register(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AccountApi->register: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**RegisterBody**](RegisterBody.md)|  | [optional] 

### Return type

[**InlineResponse200**](InlineResponse200.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **reset_password**
> reset_password(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.ResetPasswordBody() # ResetPasswordBody |  (optional)

try:
    api_instance.reset_password(body=body)
except ApiException as e:
    print("Exception when calling AccountApi->reset_password: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ResetPasswordBody**](ResetPasswordBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **reset_password_request**
> reset_password_request(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.ResetPasswordRequestBody() # ResetPasswordRequestBody |  (optional)

try:
    api_instance.reset_password_request(body=body)
except ApiException as e:
    print("Exception when calling AccountApi->reset_password_request: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**ResetPasswordRequestBody**](ResetPasswordRequestBody.md)|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **verify**
> InlineResponse2001 verify(body=body)



### Example
```python
from __future__ import print_function
import time
import zrok_api
from zrok_api.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = zrok_api.AccountApi()
body = zrok_api.VerifyBody() # VerifyBody |  (optional)

try:
    api_response = api_instance.verify(body=body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling AccountApi->verify: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**VerifyBody**](VerifyBody.md)|  | [optional] 

### Return type

[**InlineResponse2001**](InlineResponse2001.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/zrok.v1+json
 - **Accept**: application/zrok.v1+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

