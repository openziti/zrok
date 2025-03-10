# RegisterRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**register_token** | **str** |  | [optional] 
**password** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.register_request import RegisterRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RegisterRequest from a JSON string
register_request_instance = RegisterRequest.from_json(json)
# print the JSON string representation of the object
print(RegisterRequest.to_json())

# convert the object into a dict
register_request_dict = register_request_instance.to_dict()
# create an instance of RegisterRequest from a dict
register_request_from_dict = RegisterRequest.from_dict(register_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


