# CreateOrganization201Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**organization_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.create_organization201_response import CreateOrganization201Response

# TODO update the JSON string below
json = "{}"
# create an instance of CreateOrganization201Response from a JSON string
create_organization201_response_instance = CreateOrganization201Response.from_json(json)
# print the JSON string representation of the object
print(CreateOrganization201Response.to_json())

# convert the object into a dict
create_organization201_response_dict = create_organization201_response_instance.to_dict()
# create an instance of CreateOrganization201Response from a dict
create_organization201_response_from_dict = CreateOrganization201Response.from_dict(create_organization201_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


