# ListMemberships200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**memberships** | [**List[ListMemberships200ResponseMembershipsInner]**](ListMemberships200ResponseMembershipsInner.md) |  | [optional] 

## Example

```python
from zrok_api.models.list_memberships200_response import ListMemberships200Response

# TODO update the JSON string below
json = "{}"
# create an instance of ListMemberships200Response from a JSON string
list_memberships200_response_instance = ListMemberships200Response.from_json(json)
# print the JSON string representation of the object
print(ListMemberships200Response.to_json())

# convert the object into a dict
list_memberships200_response_dict = list_memberships200_response_instance.to_dict()
# create an instance of ListMemberships200Response from a dict
list_memberships200_response_from_dict = ListMemberships200Response.from_dict(list_memberships200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


