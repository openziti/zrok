# VerifyRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**register_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.verify_request import VerifyRequest

# TODO update the JSON string below
json = "{}"
# create an instance of VerifyRequest from a JSON string
verify_request_instance = VerifyRequest.from_json(json)
# print the JSON string representation of the object
print(VerifyRequest.to_json())

# convert the object into a dict
verify_request_dict = verify_request_instance.to_dict()
# create an instance of VerifyRequest from a dict
verify_request_from_dict = VerifyRequest.from_dict(verify_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


