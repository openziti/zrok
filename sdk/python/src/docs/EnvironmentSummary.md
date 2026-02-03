# EnvironmentSummary


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**host** | **str** |  | [optional] 
**address** | **str** |  | [optional] 
**remote_agent** | **bool** |  | [optional] 
**share_count** | **int** |  | [optional] 
**access_count** | **int** |  | [optional] 
**limited** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.environment_summary import EnvironmentSummary

# TODO update the JSON string below
json = "{}"
# create an instance of EnvironmentSummary from a JSON string
environment_summary_instance = EnvironmentSummary.from_json(json)
# print the JSON string representation of the object
print(EnvironmentSummary.to_json())

# convert the object into a dict
environment_summary_dict = environment_summary_instance.to_dict()
# create an instance of EnvironmentSummary from a dict
environment_summary_from_dict = EnvironmentSummary.from_dict(environment_summary_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


