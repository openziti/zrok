# Frontend


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | [optional] 
**frontend_token** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**bind_address** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**z_id** | **str** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.frontend import Frontend

# TODO update the JSON string below
json = "{}"
# create an instance of Frontend from a JSON string
frontend_instance = Frontend.from_json(json)
# print the JSON string representation of the object
print(Frontend.to_json())

# convert the object into a dict
frontend_dict = frontend_instance.to_dict()
# create an instance of Frontend from a dict
frontend_from_dict = Frontend.from_dict(frontend_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


