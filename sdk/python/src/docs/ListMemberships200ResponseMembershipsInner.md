# ListMemberships200ResponseMembershipsInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**organization_token** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**admin** | **bool** |  | [optional] 

## Example

```python
from zrok_api.models.list_memberships200_response_memberships_inner import ListMemberships200ResponseMembershipsInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListMemberships200ResponseMembershipsInner from a JSON string
list_memberships200_response_memberships_inner_instance = ListMemberships200ResponseMembershipsInner.from_json(json)
# print the JSON string representation of the object
print(ListMemberships200ResponseMembershipsInner.to_json())

# convert the object into a dict
list_memberships200_response_memberships_inner_dict = list_memberships200_response_memberships_inner_instance.to_dict()
# create an instance of ListMemberships200ResponseMembershipsInner from a dict
list_memberships200_response_memberships_inner_from_dict = ListMemberships200ResponseMembershipsInner.from_dict(list_memberships200_response_memberships_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


