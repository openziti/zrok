# RemoveNamespaceFrontendMappingRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**frontend_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.remove_namespace_frontend_mapping_request import RemoveNamespaceFrontendMappingRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoveNamespaceFrontendMappingRequest from a JSON string
remove_namespace_frontend_mapping_request_instance = RemoveNamespaceFrontendMappingRequest.from_json(json)
# print the JSON string representation of the object
print(RemoveNamespaceFrontendMappingRequest.to_json())

# convert the object into a dict
remove_namespace_frontend_mapping_request_dict = remove_namespace_frontend_mapping_request_instance.to_dict()
# create an instance of RemoveNamespaceFrontendMappingRequest from a dict
remove_namespace_frontend_mapping_request_from_dict = RemoveNamespaceFrontendMappingRequest.from_dict(remove_namespace_frontend_mapping_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


