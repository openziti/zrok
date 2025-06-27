# ListPublicFrontendsForAccount200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**public_frontends** | [**List[ListPublicFrontendsForAccount200ResponsePublicFrontendsInner]**](ListPublicFrontendsForAccount200ResponsePublicFrontendsInner.md) |  | [optional] 

## Example

```python
from zrok_api.models.list_public_frontends_for_account200_response import ListPublicFrontendsForAccount200Response

# TODO update the JSON string below
json = "{}"
# create an instance of ListPublicFrontendsForAccount200Response from a JSON string
list_public_frontends_for_account200_response_instance = ListPublicFrontendsForAccount200Response.from_json(json)
# print the JSON string representation of the object
print(ListPublicFrontendsForAccount200Response.to_json())

# convert the object into a dict
list_public_frontends_for_account200_response_dict = list_public_frontends_for_account200_response_instance.to_dict()
# create an instance of ListPublicFrontendsForAccount200Response from a dict
list_public_frontends_for_account200_response_from_dict = ListPublicFrontendsForAccount200Response.from_dict(list_public_frontends_for_account200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


