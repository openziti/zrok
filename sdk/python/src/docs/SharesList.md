# SharesList


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**shares** | [**List[ShareSummary]**](ShareSummary.md) |  | [optional] 

## Example

```python
from zrok_api.models.shares_list import SharesList

# TODO update the JSON string below
json = "{}"
# create an instance of SharesList from a JSON string
shares_list_instance = SharesList.from_json(json)
# print the JSON string representation of the object
print(SharesList.to_json())

# convert the object into a dict
shares_list_dict = shares_list_instance.to_dict()
# create an instance of SharesList from a dict
shares_list_from_dict = SharesList.from_dict(shares_list_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


