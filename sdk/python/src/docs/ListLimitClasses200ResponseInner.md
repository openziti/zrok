# ListLimitClasses200ResponseInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**label** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**environments** | **int** |  | [optional] 
**shares** | **int** |  | [optional] 
**reserved_shares** | **int** |  | [optional] 
**unique_names** | **int** |  | [optional] 
**share_frontends** | **int** |  | [optional] 
**period_minutes** | **int** |  | [optional] 
**rx_bytes** | **int** |  | [optional] 
**tx_bytes** | **int** |  | [optional] 
**total_bytes** | **int** |  | [optional] 
**limit_action** | **str** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.list_limit_classes200_response_inner import ListLimitClasses200ResponseInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListLimitClasses200ResponseInner from a JSON string
list_limit_classes200_response_inner_instance = ListLimitClasses200ResponseInner.from_json(json)
# print the JSON string representation of the object
print(ListLimitClasses200ResponseInner.to_json())

# convert the object into a dict
list_limit_classes200_response_inner_dict = list_limit_classes200_response_inner_instance.to_dict()
# create an instance of ListLimitClasses200ResponseInner from a dict
list_limit_classes200_response_inner_from_dict = ListLimitClasses200ResponseInner.from_dict(list_limit_classes200_response_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


