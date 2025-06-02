# RemoteStatus200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**shares** | [**List[RemoteStatus200ResponseSharesInner]**](RemoteStatus200ResponseSharesInner.md) |  | [optional] 
**accesses** | [**List[RemoteStatus200ResponseAccessesInner]**](RemoteStatus200ResponseAccessesInner.md) |  | [optional] 

## Example

```python
from zrok_api.models.remote_status200_response import RemoteStatus200Response

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteStatus200Response from a JSON string
remote_status200_response_instance = RemoteStatus200Response.from_json(json)
# print the JSON string representation of the object
print(RemoteStatus200Response.to_json())

# convert the object into a dict
remote_status200_response_dict = remote_status200_response_instance.to_dict()
# create an instance of RemoteStatus200Response from a dict
remote_status200_response_from_dict = RemoteStatus200Response.from_dict(remote_status200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


