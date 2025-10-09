# OverviewNamesInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**namespace_name** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 
**reserved** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.overview_names_inner import OverviewNamesInner

# TODO update the JSON string below
json = "{}"
# create an instance of OverviewNamesInner from a JSON string
overview_names_inner_instance = OverviewNamesInner.from_json(json)
# print the JSON string representation of the object
print(OverviewNamesInner.to_json())

# convert the object into a dict
overview_names_inner_dict = overview_names_inner_instance.to_dict()
# create an instance of OverviewNamesInner from a dict
overview_names_inner_from_dict = OverviewNamesInner.from_dict(overview_names_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


