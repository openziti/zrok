# AccessesList


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**accesses** | [**List[AccessSummary]**](AccessSummary.md) |  | [optional] 

## Example

```python
from zrok_api.models.accesses_list import AccessesList

# TODO update the JSON string below
json = "{}"
# create an instance of AccessesList from a JSON string
accesses_list_instance = AccessesList.from_json(json)
# print the JSON string representation of the object
print(AccessesList.to_json())

# convert the object into a dict
accesses_list_dict = accesses_list_instance.to_dict()
# create an instance of AccessesList from a dict
accesses_list_from_dict = AccessesList.from_dict(accesses_list_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


