# ShareRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**frontend_selection** | **List[str]** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**backend_proxy_endpoint** | **str** |  | [optional] 
**auth_scheme** | **str** |  | [optional] 
**auth_users** | [**List[AuthUser]**](AuthUser.md) |  | [optional] 
**oauth_provider** | **str** |  | [optional] 
**oauth_email_domains** | **List[str]** |  | [optional] 
**oauth_authorization_check_interval** | **str** |  | [optional] 
**reserved** | **bool** |  | [optional] 
**permission_mode** | **str** |  | [optional] 
**access_grants** | **List[str]** |  | [optional] 
**unique_name** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.share_request import ShareRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ShareRequest from a JSON string
share_request_instance = ShareRequest.from_json(json)
# print the JSON string representation of the object
print(ShareRequest.to_json())

# convert the object into a dict
share_request_dict = share_request_instance.to_dict()
# create an instance of ShareRequest from a dict
share_request_from_dict = ShareRequest.from_dict(share_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


