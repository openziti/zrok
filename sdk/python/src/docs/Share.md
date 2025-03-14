# Share


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**share_token** | **str** |  | [optional] 
**z_id** | **str** |  | [optional] 
**share_mode** | **str** |  | [optional] 
**backend_mode** | **str** |  | [optional] 
**frontend_selection** | **str** |  | [optional] 
**frontend_endpoint** | **str** |  | [optional] 
**backend_proxy_endpoint** | **str** |  | [optional] 
**reserved** | **bool** |  | [optional] 
**activity** | [**List[SparkDataSample]**](SparkDataSample.md) |  | [optional] 
**limited** | **bool** |  | [optional] 
**created_at** | **int** |  | [optional] 
**updated_at** | **int** |  | [optional] 

## Example

```python
from zrok_api.models.share import Share

# TODO update the JSON string below
json = "{}"
# create an instance of Share from a JSON string
share_instance = Share.from_json(json)
# print the JSON string representation of the object
print(Share.to_json())

# convert the object into a dict
share_dict = share_instance.to_dict()
# create an instance of Share from a dict
share_from_dict = Share.from_dict(share_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


