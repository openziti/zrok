# MetricsSample


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**rx** | **float** |  | [optional] 
**tx** | **float** |  | [optional] 
**timestamp** | **float** |  | [optional] 

## Example

```python
from zrok_api.models.metrics_sample import MetricsSample

# TODO update the JSON string below
json = "{}"
# create an instance of MetricsSample from a JSON string
metrics_sample_instance = MetricsSample.from_json(json)
# print the JSON string representation of the object
print(MetricsSample.to_json())

# convert the object into a dict
metrics_sample_dict = metrics_sample_instance.to_dict()
# create an instance of MetricsSample from a dict
metrics_sample_from_dict = MetricsSample.from_dict(metrics_sample_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


