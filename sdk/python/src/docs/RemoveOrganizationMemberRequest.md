# RemoveOrganizationMemberRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**organization_token** | **str** |  | [optional] 
**email** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.remove_organization_member_request import RemoveOrganizationMemberRequest

# TODO update the JSON string below
json = "{}"
# create an instance of RemoveOrganizationMemberRequest from a JSON string
remove_organization_member_request_instance = RemoveOrganizationMemberRequest.from_json(json)
# print the JSON string representation of the object
print(RemoveOrganizationMemberRequest.to_json())

# convert the object into a dict
remove_organization_member_request_dict = remove_organization_member_request_instance.to_dict()
# create an instance of RemoveOrganizationMemberRequest from a dict
remove_organization_member_request_from_dict = RemoveOrganizationMemberRequest.from_dict(remove_organization_member_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


