# PingRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.ping_request import PingRequest

# TODO update the JSON string below
json = "{}"
# create an instance of PingRequest from a JSON string
ping_request_instance = PingRequest.from_json(json)
# print the JSON string representation of the object
print(PingRequest.to_json())

# convert the object into a dict
ping_request_dict = ping_request_instance.to_dict()
# create an instance of PingRequest from a dict
ping_request_from_dict = PingRequest.from_dict(ping_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


