# ResetPasswordRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**reset_token** | **str** |  | [optional] 
**password** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.reset_password_request import ResetPasswordRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ResetPasswordRequest from a JSON string
reset_password_request_instance = ResetPasswordRequest.from_json(json)
# print the JSON string representation of the object
print(ResetPasswordRequest.to_json())

# convert the object into a dict
reset_password_request_dict = reset_password_request_instance.to_dict()
# create an instance of ResetPasswordRequest from a dict
reset_password_request_from_dict = ResetPasswordRequest.from_dict(reset_password_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


