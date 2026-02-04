# RemoteStatus200ResponseSharesInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**frontend_endpoints** | **List[str]** |  | [optional] 
**backend_endpoint** | **str** |  | [optional] 
**open** | **bool** |  | [optional] 
**status** | **str** |  | [optional] 
**failure** | [**RemoteStatus200ResponseSharesInnerFailure**](RemoteStatus200ResponseSharesInnerFailure.md) |  | [optional] 

## Example

```python
from zrok_api.models.remote_status200_response_shares_inner import RemoteStatus200ResponseSharesInner

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteStatus200ResponseSharesInner from a JSON string
remote_status200_response_shares_inner_instance = RemoteStatus200ResponseSharesInner.from_json(json)
# print the JSON string representation of the object
print(RemoteStatus200ResponseSharesInner.to_json())

# convert the object into a dict
remote_status200_response_shares_inner_dict = remote_status200_response_shares_inner_instance.to_dict()
# create an instance of RemoteStatus200ResponseSharesInner from a dict
remote_status200_response_shares_inner_from_dict = RemoteStatus200ResponseSharesInner.from_dict(remote_status200_response_shares_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


