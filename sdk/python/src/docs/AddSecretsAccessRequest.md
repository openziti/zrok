# AddSecretsAccessRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**secrets_access_identity_zid** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.add_secrets_access_request import AddSecretsAccessRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AddSecretsAccessRequest from a JSON string
add_secrets_access_request_instance = AddSecretsAccessRequest.from_json(json)
# print the JSON string representation of the object
print(AddSecretsAccessRequest.to_json())

# convert the object into a dict
add_secrets_access_request_dict = add_secrets_access_request_instance.to_dict()
# create an instance of AddSecretsAccessRequest from a dict
add_secrets_access_request_from_dict = AddSecretsAccessRequest.from_dict(add_secrets_access_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


