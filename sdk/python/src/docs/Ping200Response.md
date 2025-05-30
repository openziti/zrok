# Ping200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**version** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.ping200_response import Ping200Response

# TODO update the JSON string below
json = "{}"
# create an instance of Ping200Response from a JSON string
ping200_response_instance = Ping200Response.from_json(json)
# print the JSON string representation of the object
print(Ping200Response.to_json())

# convert the object into a dict
ping200_response_dict = ping200_response_instance.to_dict()
# create an instance of Ping200Response from a dict
ping200_response_from_dict = Ping200Response.from_dict(ping200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


