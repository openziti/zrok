# RegenerateAccountTokenRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email_address** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.regenerate_account_token_request import RegenerateAccountTokenRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RegenerateAccountTokenRequest from a JSON string
regenerate_account_token_request_instance = RegenerateAccountTokenRequest.from_json(json)
# print the JSON string representation of the object
print(RegenerateAccountTokenRequest.to_json())

# convert the object into a dict
regenerate_account_token_request_dict = regenerate_account_token_request_instance.to_dict()
# create an instance of RegenerateAccountTokenRequest from a dict
regenerate_account_token_request_from_dict = RegenerateAccountTokenRequest.from_dict(regenerate_account_token_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


