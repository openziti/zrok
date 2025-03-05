# UpdateAccessRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**bind_address** | **str** |  | [optional] 
**description** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.update_access_request import UpdateAccessRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateAccessRequest from a JSON string
update_access_request_instance = UpdateAccessRequest.from_json(json)
# print the JSON string representation of the object
print(UpdateAccessRequest.to_json())

# convert the object into a dict
update_access_request_dict = update_access_request_instance.to_dict()
# create an instance of UpdateAccessRequest from a dict
update_access_request_from_dict = UpdateAccessRequest.from_dict(update_access_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


