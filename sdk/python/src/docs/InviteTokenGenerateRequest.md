# InviteTokenGenerateRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**invite_tokens** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.invite_token_generate_request import InviteTokenGenerateRequest

# TODO update the JSON string below
json = "{}"
# create an instance of InviteTokenGenerateRequest from a JSON string
invite_token_generate_request_instance = InviteTokenGenerateRequest.from_json(json)
# print the JSON string representation of the object
print(InviteTokenGenerateRequest.to_json())

# convert the object into a dict
invite_token_generate_request_dict = invite_token_generate_request_instance.to_dict()
# create an instance of InviteTokenGenerateRequest from a dict
invite_token_generate_request_from_dict = InviteTokenGenerateRequest.from_dict(invite_token_generate_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


