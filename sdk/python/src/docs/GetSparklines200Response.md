# GetSparklines200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**sparklines** | [**List[Metrics]**](Metrics.md) |  | [optional] 

## Example

```python
from zrok_api.models.get_sparklines200_response import GetSparklines200Response

# TODO update the JSON string below
json = "{}"
# create an instance of GetSparklines200Response from a JSON string
get_sparklines200_response_instance = GetSparklines200Response.from_json(json)
# print the JSON string representation of the object
print(GetSparklines200Response.to_json())

# convert the object into a dict
get_sparklines200_response_dict = get_sparklines200_response_instance.to_dict()
# create an instance of GetSparklines200Response from a dict
get_sparklines200_response_from_dict = GetSparklines200Response.from_dict(get_sparklines200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


