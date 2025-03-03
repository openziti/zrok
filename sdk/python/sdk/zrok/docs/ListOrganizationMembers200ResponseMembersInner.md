# ListOrganizationMembers200ResponseMembersInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **str** |  | [optional] 
**admin** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.list_organization_members200_response_members_inner import ListOrganizationMembers200ResponseMembersInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListOrganizationMembers200ResponseMembersInner from a JSON string
list_organization_members200_response_members_inner_instance = ListOrganizationMembers200ResponseMembersInner.from_json(json)
# print the JSON string representation of the object
print(ListOrganizationMembers200ResponseMembersInner.to_json())

# convert the object into a dict
list_organization_members200_response_members_inner_dict = list_organization_members200_response_members_inner_instance.to_dict()
# create an instance of ListOrganizationMembers200ResponseMembersInner from a dict
list_organization_members200_response_members_inner_from_dict = ListOrganizationMembers200ResponseMembersInner.from_dict(list_organization_members200_response_members_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


