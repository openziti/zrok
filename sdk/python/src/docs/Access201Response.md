# Access201Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.access201_response import Access201Response

# TODO update the JSON string below
json = "{}"
# create an instance of Access201Response from a JSON string
access201_response_instance = Access201Response.from_json(json)
# print the JSON string representation of the object
print(Access201Response.to_json())

# convert the object into a dict
access201_response_dict = access201_response_instance.to_dict()
# create an instance of Access201Response from a dict
access201_response_from_dict = Access201Response.from_dict(access201_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


