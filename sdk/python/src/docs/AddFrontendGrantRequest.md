# AddFrontendGrantRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**email** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.add_frontend_grant_request import AddFrontendGrantRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AddFrontendGrantRequest from a JSON string
add_frontend_grant_request_instance = AddFrontendGrantRequest.from_json(json)
# print the JSON string representation of the object
print(AddFrontendGrantRequest.to_json())

# convert the object into a dict
add_frontend_grant_request_dict = add_frontend_grant_request_instance.to_dict()
# create an instance of AddFrontendGrantRequest from a dict
add_frontend_grant_request_from_dict = AddFrontendGrantRequest.from_dict(add_frontend_grant_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


