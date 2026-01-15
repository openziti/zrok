# MfaChallenge200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**challenge_token** | **str** |  | [optional] 
**expires_at** | **datetime** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_challenge200_response import MfaChallenge200Response

# TODO update the JSON string below
json = "{}"
# create an instance of MfaChallenge200Response from a JSON string
mfa_challenge200_response_instance = MfaChallenge200Response.from_json(json)
# print the JSON string representation of the object
print(MfaChallenge200Response.to_json())

# convert the object into a dict
mfa_challenge200_response_dict = mfa_challenge200_response_instance.to_dict()
# create an instance of MfaChallenge200Response from a dict
mfa_challenge200_response_from_dict = MfaChallenge200Response.from_dict(mfa_challenge200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


