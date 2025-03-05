# EnvironmentAndResources


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**environment** | [**Environment**](Environment.md) |  | [optional] 
**frontends** | [**List[Frontend]**](Frontend.md) |  | [optional] 
**shares** | [**List[Share]**](Share.md) |  | [optional] 

## Example

```python
from zrok_api.models.environment_and_resources import EnvironmentAndResources

# TODO update the JSON string below
json = "{}"
# create an instance of EnvironmentAndResources from a JSON string
environment_and_resources_instance = EnvironmentAndResources.from_json(json)
# print the JSON string representation of the object
print(EnvironmentAndResources.to_json())

# convert the object into a dict
environment_and_resources_dict = environment_and_resources_instance.to_dict()
# create an instance of EnvironmentAndResources from a dict
environment_and_resources_from_dict = EnvironmentAndResources.from_dict(environment_and_resources_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


