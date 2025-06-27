# DeleteIdentityRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**z_id** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.delete_identity_request import DeleteIdentityRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DeleteIdentityRequest from a JSON string
delete_identity_request_instance = DeleteIdentityRequest.from_json(json)
# print the JSON string representation of the object
print(DeleteIdentityRequest.to_json())

# convert the object into a dict
delete_identity_request_dict = delete_identity_request_instance.to_dict()
# create an instance of DeleteIdentityRequest from a dict
delete_identity_request_from_dict = DeleteIdentityRequest.from_dict(delete_identity_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


