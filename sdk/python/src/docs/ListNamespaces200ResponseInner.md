# ListNamespaces200ResponseInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**open** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.list_namespaces200_response_inner import ListNamespaces200ResponseInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListNamespaces200ResponseInner from a JSON string
list_namespaces200_response_inner_instance = ListNamespaces200ResponseInner.from_json(json)
# print the JSON string representation of the object
print(ListNamespaces200ResponseInner.to_json())

# convert the object into a dict
list_namespaces200_response_inner_dict = list_namespaces200_response_inner_instance.to_dict()
# create an instance of ListNamespaces200ResponseInner from a dict
list_namespaces200_response_inner_from_dict = ListNamespaces200ResponseInner.from_dict(list_namespaces200_response_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


