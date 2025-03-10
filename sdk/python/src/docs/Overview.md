# Overview


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**account_limited** | **bool** |  | [optional] 
**environments** | [**List[EnvironmentAndResources]**](EnvironmentAndResources.md) |  | [optional] 

## Example

```python
from zrok_api.models.overview import Overview

# TODO update the JSON string below
json = "{}"
# create an instance of Overview from a JSON string
overview_instance = Overview.from_json(json)
# print the JSON string representation of the object
print(Overview.to_json())

# convert the object into a dict
overview_dict = overview_instance.to_dict()
# create an instance of Overview from a dict
overview_from_dict = Overview.from_dict(overview_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


