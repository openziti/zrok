# EnableRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**description** | **str** |  | [optional] 
**host** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.enable_request import EnableRequest

# TODO update the JSON string below
json = "{}"
# create an instance of EnableRequest from a JSON string
enable_request_instance = EnableRequest.from_json(json)
# print the JSON string representation of the object
print(EnableRequest.to_json())

# convert the object into a dict
enable_request_dict = enable_request_instance.to_dict()
# create an instance of EnableRequest from a dict
enable_request_from_dict = EnableRequest.from_dict(enable_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


