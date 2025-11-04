# ShareSummary


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**share_token** | **str** |  | [optional] 
**z_id** | **str** |  | [optional] 
**env_zid** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**frontend_endpoints** | **List[str]** |  | [optional] 
**target** | **str** |  | [optional] 
**limited** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.share_summary import ShareSummary

# TODO update the JSON string below
json = "{}"
# create an instance of ShareSummary from a JSON string
share_summary_instance = ShareSummary.from_json(json)
# print the JSON string representation of the object
print(ShareSummary.to_json())

# convert the object into a dict
share_summary_dict = share_summary_instance.to_dict()
# create an instance of ShareSummary from a dict
share_summary_from_dict = ShareSummary.from_dict(share_summary_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


