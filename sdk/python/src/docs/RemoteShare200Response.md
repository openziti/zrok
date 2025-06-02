# RemoteShare200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **str** |  | [optional] 
**frontend_endpoints** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.remote_share200_response import RemoteShare200Response

# TODO update the JSON string below
json = "{}"
# create an instance of RemoteShare200Response from a JSON string
remote_share200_response_instance = RemoteShare200Response.from_json(json)
# print the JSON string representation of the object
print(RemoteShare200Response.to_json())

# convert the object into a dict
remote_share200_response_dict = remote_share200_response_instance.to_dict()
# create an instance of RemoteShare200Response from a dict
remote_share200_response_from_dict = RemoteShare200Response.from_dict(remote_share200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


