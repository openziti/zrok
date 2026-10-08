# GetName200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**namespace_token** | **str** |  | 
**namespace_name** | **str** |  | 
**name** | **str** |  | 
**account_email** | **str** |  | 
**share_token** | **str** |  | 
**reserved** | **bool** |  | 
**created_at** | **int** |  | 

## Example

```python
from zrok_api.models.get_name200_response import GetName200Response

# TODO update the JSON string below
json = "{}"
# create an instance of GetName200Response from a JSON string
get_name200_response_instance = GetName200Response.from_json(json)
# print the JSON string representation of the object
print(GetName200Response.to_json())

# convert the object into a dict
get_name200_response_dict = get_name200_response_instance.to_dict()
# create an instance of GetName200Response from a dict
get_name200_response_from_dict = GetName200Response.from_dict(get_name200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


