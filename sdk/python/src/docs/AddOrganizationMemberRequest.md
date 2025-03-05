# AddOrganizationMemberRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**organization_token** | **str** |  | [optional] 
**email** | **str** |  | [optional] 
**admin** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.add_organization_member_request import AddOrganizationMemberRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AddOrganizationMemberRequest from a JSON string
add_organization_member_request_instance = AddOrganizationMemberRequest.from_json(json)
# print the JSON string representation of the object
print(AddOrganizationMemberRequest.to_json())

# convert the object into a dict
add_organization_member_request_dict = add_organization_member_request_instance.to_dict()
# create an instance of AddOrganizationMemberRequest from a dict
add_organization_member_request_from_dict = AddOrganizationMemberRequest.from_dict(add_organization_member_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


