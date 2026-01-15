# MfaDisableRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**password** | **str** |  | [optional] 
**code** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.mfa_disable_request import MfaDisableRequest

# TODO update the JSON string below
json = "{}"
# create an instance of MfaDisableRequest from a JSON string
mfa_disable_request_instance = MfaDisableRequest.from_json(json)
# print the JSON string representation of the object
print(MfaDisableRequest.to_json())

# convert the object into a dict
mfa_disable_request_dict = mfa_disable_request_instance.to_dict()
# create an instance of MfaDisableRequest from a dict
mfa_disable_request_from_dict = MfaDisableRequest.from_dict(mfa_disable_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


