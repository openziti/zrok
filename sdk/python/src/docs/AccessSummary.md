# AccessSummary


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**frontend_token** | **str** |  | [optional] 
**env_zid** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**bind_address** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**limited** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.access_summary import AccessSummary

# TODO update the JSON string below
json = "{}"
# create an instance of AccessSummary from a JSON string
access_summary_instance = AccessSummary.from_json(json)
# print the JSON string representation of the object
print(AccessSummary.to_json())

# convert the object into a dict
access_summary_dict = access_summary_instance.to_dict()
# create an instance of AccessSummary from a dict
access_summary_from_dict = AccessSummary.from_dict(access_summary_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


