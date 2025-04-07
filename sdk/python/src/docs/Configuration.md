# Configuration


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**version** | **str** |  | [optional] 
**tou_link** | **str** |  | [optional] 
**new_account_link** | **str** |  | [optional] 
**invites_open** | **bool** |  | [optional] 
**requires_invite_token** | **bool** |  | [optional] 
**invite_token_contact** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.configuration import Configuration

# TODO update the JSON string below
json = "{}"
# create an instance of Configuration from a JSON string
configuration_instance = Configuration.from_json(json)
# print the JSON string representation of the object
print(Configuration.to_json())

# convert the object into a dict
configuration_dict = configuration_instance.to_dict()
# create an instance of Configuration from a dict
configuration_from_dict = Configuration.from_dict(configuration_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


