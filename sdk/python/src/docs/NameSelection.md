# NameSelection


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | [optional] 
**name** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.name_selection import NameSelection

# TODO update the JSON string below
json = "{}"
# create an instance of NameSelection from a JSON string
name_selection_instance = NameSelection.from_json(json)
# print the JSON string representation of the object
print(NameSelection.to_json())

# convert the object into a dict
name_selection_dict = name_selection_instance.to_dict()
# create an instance of NameSelection from a dict
name_selection_from_dict = NameSelection.from_dict(name_selection_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


