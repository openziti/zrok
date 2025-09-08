# Unshare12Request


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.unshare12_request import Unshare12Request

# TODO update the JSON string below
json = "{}"
# create an instance of Unshare12Request from a JSON string
unshare12_request_instance = Unshare12Request.from_json(json)
# print the JSON string representation of the object
print(Unshare12Request.to_json())

# convert the object into a dict
unshare12_request_dict = unshare12_request_instance.to_dict()
# create an instance of Unshare12Request from a dict
unshare12_request_from_dict = Unshare12Request.from_dict(unshare12_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


