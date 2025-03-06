# ClientVersionCheckRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**client_version** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.client_version_check_request import ClientVersionCheckRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ClientVersionCheckRequest from a JSON string
client_version_check_request_instance = ClientVersionCheckRequest.from_json(json)
# print the JSON string representation of the object
print(ClientVersionCheckRequest.to_json())

# convert the object into a dict
client_version_check_request_dict = client_version_check_request_instance.to_dict()
# create an instance of ClientVersionCheckRequest from a dict
client_version_check_request_from_dict = ClientVersionCheckRequest.from_dict(client_version_check_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


