# ListOrganizationMembers200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**members** | [**List[ListOrganizationMembers200ResponseMembersInner]**](ListOrganizationMembers200ResponseMembersInner.md) |  | [optional] 

## Example

```python
from zrok_api.models.list_organization_members200_response import ListOrganizationMembers200Response

# TODO update the JSON string below
json = "{}"
# create an instance of ListOrganizationMembers200Response from a JSON string
list_organization_members200_response_instance = ListOrganizationMembers200Response.from_json(json)
# print the JSON string representation of the object
print(ListOrganizationMembers200Response.to_json())

# convert the object into a dict
list_organization_members200_response_dict = list_organization_members200_response_instance.to_dict()
# create an instance of ListOrganizationMembers200Response from a dict
list_organization_members200_response_from_dict = ListOrganizationMembers200Response.from_dict(list_organization_members200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


