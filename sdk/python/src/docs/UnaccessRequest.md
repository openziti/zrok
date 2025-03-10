# UnaccessRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**env_zid** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.unaccess_request import UnaccessRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UnaccessRequest from a JSON string
unaccess_request_instance = UnaccessRequest.from_json(json)
# print the JSON string representation of the object
print(UnaccessRequest.to_json())

# convert the object into a dict
unaccess_request_dict = unaccess_request_instance.to_dict()
# create an instance of UnaccessRequest from a dict
unaccess_request_from_dict = UnaccessRequest.from_dict(unaccess_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


