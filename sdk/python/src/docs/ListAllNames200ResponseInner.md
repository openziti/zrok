# ListAllNames200ResponseInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**namespace_name** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**reserved** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.list_all_names200_response_inner import ListAllNames200ResponseInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListAllNames200ResponseInner from a JSON string
list_all_names200_response_inner_instance = ListAllNames200ResponseInner.from_json(json)
# print the JSON string representation of the object
print(ListAllNames200ResponseInner.to_json())

# convert the object into a dict
list_all_names200_response_inner_dict = list_all_names200_response_inner_instance.to_dict()
# create an instance of ListAllNames200ResponseInner from a dict
list_all_names200_response_inner_from_dict = ListAllNames200ResponseInner.from_dict(list_all_names200_response_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


