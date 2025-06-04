# RemoteShareRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**token** | **str** |  | [optional] 
**target** | **str** |  | [optional] 
**basic_auth** | **List[str]** |  | [optional] 
**frontend_selection** | **List[str]** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**insecure** | **bool** |  | [optional] 
**oauth_provider** | **str** |  | [optional] 
**oauth_email_address_patterns** | **List[str]** |  | [optional] 
**oauth_check_interval** | **str** |  | [optional] 
**open** | **bool** |  | [optional] 
**access_grants** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.remote_share_request import RemoteShareRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteShareRequest from a JSON string
remote_share_request_instance = RemoteShareRequest.from_json(json)
# print the JSON string representation of the object
print(RemoteShareRequest.to_json())

# convert the object into a dict
remote_share_request_dict = remote_share_request_instance.to_dict()
# create an instance of RemoteShareRequest from a dict
remote_share_request_from_dict = RemoteShareRequest.from_dict(remote_share_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


