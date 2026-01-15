# MfaAuthenticateRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**pending_token** | **str** |  | [optional] 
**code** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_authenticate_request import MfaAuthenticateRequest

# TODO update the JSON string below
json = "{}"
# create an instance of MfaAuthenticateRequest from a JSON string
mfa_authenticate_request_instance = MfaAuthenticateRequest.from_json(json)
# print the JSON string representation of the object
print(MfaAuthenticateRequest.to_json())

# convert the object into a dict
mfa_authenticate_request_dict = mfa_authenticate_request_instance.to_dict()
# create an instance of MfaAuthenticateRequest from a dict
mfa_authenticate_request_from_dict = MfaAuthenticateRequest.from_dict(mfa_authenticate_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


