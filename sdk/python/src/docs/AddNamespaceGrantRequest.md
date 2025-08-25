# AddNamespaceGrantRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**email** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.add_namespace_grant_request import AddNamespaceGrantRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AddNamespaceGrantRequest from a JSON string
add_namespace_grant_request_instance = AddNamespaceGrantRequest.from_json(json)
# print the JSON string representation of the object
print(AddNamespaceGrantRequest.to_json())

# convert the object into a dict
add_namespace_grant_request_dict = add_namespace_grant_request_instance.to_dict()
# create an instance of AddNamespaceGrantRequest from a dict
add_namespace_grant_request_from_dict = AddNamespaceGrantRequest.from_dict(add_namespace_grant_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


