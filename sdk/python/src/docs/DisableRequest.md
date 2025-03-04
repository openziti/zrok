# DisableRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**identity** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.disable_request import DisableRequest

# TODO update the JSON string below
json = "{}"
# create an instance of DisableRequest from a JSON string
disable_request_instance = DisableRequest.from_json(json)
# print the JSON string representation of the object
print(DisableRequest.to_json())

# convert the object into a dict
disable_request_dict = disable_request_instance.to_dict()
# create an instance of DisableRequest from a dict
disable_request_from_dict = DisableRequest.from_dict(disable_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


