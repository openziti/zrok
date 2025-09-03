# NamespaceSelection


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.namespace_selection import NamespaceSelection

# TODO update the JSON string below
json = "{}"
# create an instance of NamespaceSelection from a JSON string
namespace_selection_instance = NamespaceSelection.from_json(json)
# print the JSON string representation of the object
print(NamespaceSelection.to_json())

# convert the object into a dict
namespace_selection_dict = namespace_selection_instance.to_dict()
# create an instance of NamespaceSelection from a dict
namespace_selection_from_dict = NamespaceSelection.from_dict(namespace_selection_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


