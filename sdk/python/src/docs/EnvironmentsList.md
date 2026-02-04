# EnvironmentsList


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**environments** | [**List[EnvironmentSummary]**](EnvironmentSummary.md) |  | [optional] 

## Example

```python
from zrok_api.models.environments_list import EnvironmentsList

# TODO update the JSON string below
json = "{}"
# create an instance of EnvironmentsList from a JSON string
environments_list_instance = EnvironmentsList.from_json(json)
# print the JSON string representation of the object
print(EnvironmentsList.to_json())

# convert the object into a dict
environments_list_dict = environments_list_instance.to_dict()
# create an instance of EnvironmentsList from a dict
environments_list_from_dict = EnvironmentsList.from_dict(environments_list_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


