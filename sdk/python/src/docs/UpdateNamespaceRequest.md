# UpdateNamespaceRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**open** | **bool** |  | [optional] 
**open_set** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.update_namespace_request import UpdateNamespaceRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateNamespaceRequest from a JSON string
update_namespace_request_instance = UpdateNamespaceRequest.from_json(json)
# print the JSON string representation of the object
print(UpdateNamespaceRequest.to_json())

# convert the object into a dict
update_namespace_request_dict = update_namespace_request_instance.to_dict()
# create an instance of UpdateNamespaceRequest from a dict
update_namespace_request_from_dict = UpdateNamespaceRequest.from_dict(update_namespace_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


