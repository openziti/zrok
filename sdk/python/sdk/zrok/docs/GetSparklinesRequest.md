# GetSparklinesRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**account** | **bool** |  | [optional] 
**environments** | **List[str]** |  | [optional] 
**shares** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.get_sparklines_request import GetSparklinesRequest

# TODO update the JSON string below
json = "{}"
# create an instance of GetSparklinesRequest from a JSON string
get_sparklines_request_instance = GetSparklinesRequest.from_json(json)
# print the JSON string representation of the object
print(GetSparklinesRequest.to_json())

# convert the object into a dict
get_sparklines_request_dict = get_sparklines_request_instance.to_dict()
# create an instance of GetSparklinesRequest from a dict
get_sparklines_request_from_dict = GetSparklinesRequest.from_dict(get_sparklines_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


