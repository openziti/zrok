# UpdateShareRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**share_token** | **str** |  | [optional] 
**add_access_grants** | **List[str]** |  | [optional] 
**remove_access_grants** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.update_share_request import UpdateShareRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UpdateShareRequest from a JSON string
update_share_request_instance = UpdateShareRequest.from_json(json)
# print the JSON string representation of the object
print(UpdateShareRequest.to_json())

# convert the object into a dict
update_share_request_dict = update_share_request_instance.to_dict()
# create an instance of UpdateShareRequest from a dict
update_share_request_from_dict = UpdateShareRequest.from_dict(update_share_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


