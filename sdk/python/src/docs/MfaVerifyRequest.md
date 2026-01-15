# MfaVerifyRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_verify_request import MfaVerifyRequest

# TODO update the JSON string below
json = "{}"
# create an instance of MfaVerifyRequest from a JSON string
mfa_verify_request_instance = MfaVerifyRequest.from_json(json)
# print the JSON string representation of the object
print(MfaVerifyRequest.to_json())

# convert the object into a dict
mfa_verify_request_dict = mfa_verify_request_instance.to_dict()
# create an instance of MfaVerifyRequest from a dict
mfa_verify_request_from_dict = MfaVerifyRequest.from_dict(mfa_verify_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


