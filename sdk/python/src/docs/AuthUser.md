# AuthUser


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**username** | **str** |  | [optional] 
**password** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.auth_user import AuthUser

# TODO update the JSON string below
json = "{}"
# create an instance of AuthUser from a JSON string
auth_user_instance = AuthUser.from_json(json)
# print the JSON string representation of the object
print(AuthUser.to_json())

# convert the object into a dict
auth_user_dict = auth_user_instance.to_dict()
# create an instance of AuthUser from a dict
auth_user_from_dict = AuthUser.from_dict(auth_user_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


