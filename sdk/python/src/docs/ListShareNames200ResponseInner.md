# ListShareNames200ResponseInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** |  | [optional] 
**created_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.list_share_names200_response_inner import ListShareNames200ResponseInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListShareNames200ResponseInner from a JSON string
list_share_names200_response_inner_instance = ListShareNames200ResponseInner.from_json(json)
# print the JSON string representation of the object
print(ListShareNames200ResponseInner.to_json())

# convert the object into a dict
list_share_names200_response_inner_dict = list_share_names200_response_inner_instance.to_dict()
# create an instance of ListShareNames200ResponseInner from a dict
list_share_names200_response_inner_from_dict = ListShareNames200ResponseInner.from_dict(list_share_names200_response_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


