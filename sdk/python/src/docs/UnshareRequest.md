# UnshareRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 
**reserved** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.unshare_request import UnshareRequest

# TODO update the JSON string below
json = "{}"
# create an instance of UnshareRequest from a JSON string
unshare_request_instance = UnshareRequest.from_json(json)
# print the JSON string representation of the object
print(UnshareRequest.to_json())

# convert the object into a dict
unshare_request_dict = unshare_request_instance.to_dict()
# create an instance of UnshareRequest from a dict
unshare_request_from_dict = UnshareRequest.from_dict(unshare_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


