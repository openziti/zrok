# AddNamespaceFrontendMappingRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**frontend_token** | **str** |  | [optional] 
**is_default** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.add_namespace_frontend_mapping_request import AddNamespaceFrontendMappingRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AddNamespaceFrontendMappingRequest from a JSON string
add_namespace_frontend_mapping_request_instance = AddNamespaceFrontendMappingRequest.from_json(json)
# print the JSON string representation of the object
print(AddNamespaceFrontendMappingRequest.to_json())

# convert the object into a dict
add_namespace_frontend_mapping_request_dict = add_namespace_frontend_mapping_request_instance.to_dict()
# create an instance of AddNamespaceFrontendMappingRequest from a dict
add_namespace_frontend_mapping_request_from_dict = AddNamespaceFrontendMappingRequest.from_dict(add_namespace_frontend_mapping_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


