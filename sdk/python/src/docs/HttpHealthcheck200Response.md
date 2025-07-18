# HttpHealthcheck200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**healthy** | **bool** |  | [optional] 
**error** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.http_healthcheck200_response import HttpHealthcheck200Response

# TODO update the JSON string below
json = "{}"
# create an instance of HttpHealthcheck200Response from a JSON string
http_healthcheck200_response_instance = HttpHealthcheck200Response.from_json(json)
# print the JSON string representation of the object
print(HttpHealthcheck200Response.to_json())

# convert the object into a dict
http_healthcheck200_response_dict = http_healthcheck200_response_instance.to_dict()
# create an instance of HttpHealthcheck200Response from a dict
http_healthcheck200_response_from_dict = HttpHealthcheck200Response.from_dict(http_healthcheck200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


