# CreateShareNameRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.create_share_name_request import CreateShareNameRequest

# TODO update the JSON string below
json = "{}"
# create an instance of CreateShareNameRequest from a JSON string
create_share_name_request_instance = CreateShareNameRequest.from_json(json)
# print the JSON string representation of the object
print(CreateShareNameRequest.to_json())

# convert the object into a dict
create_share_name_request_dict = create_share_name_request_instance.to_dict()
# create an instance of CreateShareNameRequest from a dict
create_share_name_request_from_dict = CreateShareNameRequest.from_dict(create_share_name_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


