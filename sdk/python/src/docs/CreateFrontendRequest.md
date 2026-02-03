# CreateFrontendRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**z_id** | **str** |  | [optional] 
**url_template** | **str** |  | [optional] 
**public_name** | **str** |  | [optional] 
**permission_mode** | **str** |  | [optional] 
**dynamic** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.create_frontend_request import CreateFrontendRequest

# TODO update the JSON string below
json = "{}"
# create an instance of CreateFrontendRequest from a JSON string
create_frontend_request_instance = CreateFrontendRequest.from_json(json)
# print the JSON string representation of the object
print(CreateFrontendRequest.to_json())

# convert the object into a dict
create_frontend_request_dict = create_frontend_request_instance.to_dict()
# create an instance of CreateFrontendRequest from a dict
create_frontend_request_from_dict = CreateFrontendRequest.from_dict(create_frontend_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


