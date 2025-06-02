# RemoteUnaccessRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**frontend_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.remote_unaccess_request import RemoteUnaccessRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteUnaccessRequest from a JSON string
remote_unaccess_request_instance = RemoteUnaccessRequest.from_json(json)
# print the JSON string representation of the object
print(RemoteUnaccessRequest.to_json())

# convert the object into a dict
remote_unaccess_request_dict = remote_unaccess_request_instance.to_dict()
# create an instance of RemoteUnaccessRequest from a dict
remote_unaccess_request_from_dict = RemoteUnaccessRequest.from_dict(remote_unaccess_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


