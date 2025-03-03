# CreateIdentity201Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**identity** | **str** |  | [optional] 
**cfg** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.create_identity201_response import CreateIdentity201Response

# TODO update the JSON string below
json = "{}"
# create an instance of CreateIdentity201Response from a JSON string
create_identity201_response_instance = CreateIdentity201Response.from_json(json)
# print the JSON string representation of the object
print(CreateIdentity201Response.to_json())

# convert the object into a dict
create_identity201_response_dict = create_identity201_response_instance.to_dict()
# create an instance of CreateIdentity201Response from a dict
create_identity201_response_from_dict = CreateIdentity201Response.from_dict(create_identity201_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


