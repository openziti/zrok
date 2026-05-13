# LimitClass


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
from zrok_api.models.limit_class import LimitClass

# TODO update the JSON string below
json = "{}"
# create an instance of LimitClass from a JSON string
limit_class_instance = LimitClass.from_json(json)
# print the JSON string representation of the object
print(LimitClass.to_json())

# convert the object into a dict
limit_class_dict = limit_class_instance.to_dict()
# create an instance of LimitClass from a dict
limit_class_from_dict = LimitClass.from_dict(limit_class_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


