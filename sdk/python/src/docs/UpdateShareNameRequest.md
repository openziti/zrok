# UpdateShareNameRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 
**reserved** | **bool** | whether the name should be reserved (true) or released (false) | [optional] 

## Example

```python
from zrok_api.models.update_share_name_request import UpdateShareNameRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateShareNameRequest from a JSON string
update_share_name_request_instance = UpdateShareNameRequest.from_json(json)
# print the JSON string representation of the object
print(UpdateShareNameRequest.to_json())

# convert the object into a dict
update_share_name_request_dict = update_share_name_request_instance.to_dict()
# create an instance of UpdateShareNameRequest from a dict
update_share_name_request_from_dict = UpdateShareNameRequest.from_dict(update_share_name_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


