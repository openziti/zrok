# ListFrontends200ResponseInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_token** | **str** |  | [optional] 
**z_id** | **str** |  | [optional] 
**url_template** | **str** |  | [optional] 
**public_name** | **str** |  | [optional] 
**permission_mode** | **str** |  | [optional] 
**dynamic** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.list_frontends200_response_inner import ListFrontends200ResponseInner

# TODO update the JSON string below
json = "{}"
# create an instance of ListFrontends200ResponseInner from a JSON string
list_frontends200_response_inner_instance = ListFrontends200ResponseInner.from_json(json)
# print the JSON string representation of the object
print(ListFrontends200ResponseInner.to_json())

# convert the object into a dict
list_frontends200_response_inner_dict = list_frontends200_response_inner_instance.to_dict()
# create an instance of ListFrontends200ResponseInner from a dict
list_frontends200_response_inner_from_dict = ListFrontends200ResponseInner.from_dict(list_frontends200_response_inner_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


