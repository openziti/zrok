# UpdateFrontendRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**public_name** | **str** |  | [optional] 
**url_template** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.update_frontend_request import UpdateFrontendRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateFrontendRequest from a JSON string
update_frontend_request_instance = UpdateFrontendRequest.from_json(json)
# print the JSON string representation of the object
print(UpdateFrontendRequest.to_json())

# convert the object into a dict
update_frontend_request_dict = update_frontend_request_instance.to_dict()
# create an instance of UpdateFrontendRequest from a dict
update_frontend_request_from_dict = UpdateFrontendRequest.from_dict(update_frontend_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


