# ShareRequest12


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**namespace_selections** | [**List[NamespaceSelection]**](NamespaceSelection.md) |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**target** | **str** |  | [optional] 
**auth_scheme** | **str** |  | [optional] 
**basic_auth_users** | [**List[AuthUser]**](AuthUser.md) |  | [optional] 
**oauth_provider** | **str** |  | [optional] 
**oauth_email_domains** | **List[str]** |  | [optional] 
**oauth_refresh_interval** | **str** |  | [optional] 
**permission_mode** | **str** |  | [optional] 
**access_grants** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.share_request12 import ShareRequest12

# TODO update the JSON string below
json = "{}"
# create an instance of ShareRequest12 from a JSON string
share_request12_instance = ShareRequest12.from_json(json)
# print the JSON string representation of the object
print(ShareRequest12.to_json())

# convert the object into a dict
share_request12_dict = share_request12_instance.to_dict()
# create an instance of ShareRequest12 from a dict
share_request12_from_dict = ShareRequest12.from_dict(share_request12_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


