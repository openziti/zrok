# MfaVerify200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**recovery_codes** | **List[str]** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_verify200_response import MfaVerify200Response

# TODO update the JSON string below
json = "{}"
# create an instance of MfaVerify200Response from a JSON string
mfa_verify200_response_instance = MfaVerify200Response.from_json(json)
# print the JSON string representation of the object
print(MfaVerify200Response.to_json())

# convert the object into a dict
mfa_verify200_response_dict = mfa_verify200_response_instance.to_dict()
# create an instance of MfaVerify200Response from a dict
mfa_verify200_response_from_dict = MfaVerify200Response.from_dict(mfa_verify200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


