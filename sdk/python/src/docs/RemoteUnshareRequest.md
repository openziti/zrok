# RemoteUnshareRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.remote_unshare_request import RemoteUnshareRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteUnshareRequest from a JSON string
remote_unshare_request_instance = RemoteUnshareRequest.from_json(json)
# print the JSON string representation of the object
print(RemoteUnshareRequest.to_json())

# convert the object into a dict
remote_unshare_request_dict = remote_unshare_request_instance.to_dict()
# create an instance of RemoteUnshareRequest from a dict
remote_unshare_request_from_dict = RemoteUnshareRequest.from_dict(remote_unshare_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


