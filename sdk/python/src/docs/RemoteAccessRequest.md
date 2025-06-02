# RemoteAccessRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**token** | **str** |  | [optional] 
**bind_address** | **str** |  | [optional] 
**auto_mode** | **bool** |  | [optional] 
**auto_address** | **str** |  | [optional] 
**auto_start_port** | **int** |  | [optional] 
**auto_end_port** | **int** |  | [optional] 
**response_headers** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.remote_access_request import RemoteAccessRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteAccessRequest from a JSON string
remote_access_request_instance = RemoteAccessRequest.from_json(json)
# print the JSON string representation of the object
print(RemoteAccessRequest.to_json())

# convert the object into a dict
remote_access_request_dict = remote_access_request_instance.to_dict()
# create an instance of RemoteAccessRequest from a dict
remote_access_request_from_dict = RemoteAccessRequest.from_dict(remote_access_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


