# ChangePasswordRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **str** |  | [optional] 
**old_password** | **str** |  | [optional] 
**new_password** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.change_password_request import ChangePasswordRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ChangePasswordRequest from a JSON string
change_password_request_instance = ChangePasswordRequest.from_json(json)
# print the JSON string representation of the object
print(ChangePasswordRequest.to_json())

# convert the object into a dict
change_password_request_dict = change_password_request_instance.to_dict()
# create an instance of ChangePasswordRequest from a dict
change_password_request_from_dict = ChangePasswordRequest.from_dict(change_password_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


