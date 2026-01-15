# MfaStatus200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enabled** | **bool** |  | [optional] 
**recovery_codes_remaining** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_status200_response import MfaStatus200Response

# TODO update the JSON string below
json = "{}"
# create an instance of MfaStatus200Response from a JSON string
mfa_status200_response_instance = MfaStatus200Response.from_json(json)
# print the JSON string representation of the object
print(MfaStatus200Response.to_json())

# convert the object into a dict
mfa_status200_response_dict = mfa_status200_response_instance.to_dict()
# create an instance of MfaStatus200Response from a dict
mfa_status200_response_from_dict = MfaStatus200Response.from_dict(mfa_status200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


