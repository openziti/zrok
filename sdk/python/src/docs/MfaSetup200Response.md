# MfaSetup200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**secret** | **str** |  | [optional] 
**qr_code** | **str** |  | [optional] 
**provisioning_uri** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_setup200_response import MfaSetup200Response

# TODO update the JSON string below
json = "{}"
# create an instance of MfaSetup200Response from a JSON string
mfa_setup200_response_instance = MfaSetup200Response.from_json(json)
# print the JSON string representation of the object
print(MfaSetup200Response.to_json())

# convert the object into a dict
mfa_setup200_response_dict = mfa_setup200_response_instance.to_dict()
# create an instance of MfaSetup200Response from a dict
mfa_setup200_response_from_dict = MfaSetup200Response.from_dict(mfa_setup200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


