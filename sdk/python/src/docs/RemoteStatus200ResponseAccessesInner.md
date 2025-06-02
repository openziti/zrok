# RemoteStatus200ResponseAccessesInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**token** | **str** |  | [optional] 
**bind_address** | **str** |  | [optional] 
**response_headers** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.remote_status200_response_accesses_inner import RemoteStatus200ResponseAccessesInner

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteStatus200ResponseAccessesInner from a JSON string
remote_status200_response_accesses_inner_instance = RemoteStatus200ResponseAccessesInner.from_json(json)
# print the JSON string representation of the object
print(RemoteStatus200ResponseAccessesInner.to_json())

# convert the object into a dict
remote_status200_response_accesses_inner_dict = remote_status200_response_accesses_inner_instance.to_dict()
# create an instance of RemoteStatus200ResponseAccessesInner from a dict
remote_status200_response_accesses_inner_from_dict = RemoteStatus200ResponseAccessesInner.from_dict(remote_status200_response_accesses_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


