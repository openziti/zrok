# RemoteStatus200ResponseSharesInnerFailure


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **str** |  | [optional] 
**count** | **int** |  | [optional] 
**last_error** | **str** |  | [optional] 
**next_retry** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.remote_status200_response_shares_inner_failure import RemoteStatus200ResponseSharesInnerFailure

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteStatus200ResponseSharesInnerFailure from a JSON string
remote_status200_response_shares_inner_failure_instance = RemoteStatus200ResponseSharesInnerFailure.from_json(json)
# print the JSON string representation of the object
print(RemoteStatus200ResponseSharesInnerFailure.to_json())

# convert the object into a dict
remote_status200_response_shares_inner_failure_dict = remote_status200_response_shares_inner_failure_instance.to_dict()
# create an instance of RemoteStatus200ResponseSharesInnerFailure from a dict
remote_status200_response_shares_inner_failure_from_dict = RemoteStatus200ResponseSharesInnerFailure.from_dict(remote_status200_response_shares_inner_failure_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


