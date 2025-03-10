# SparkDataSample


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**rx** | **float** |  | [optional] 
**tx** | **float** |  | [optional] 

## Example

```python
from zrok_api.models.spark_data_sample import SparkDataSample

# TODO update the JSON string below
json = "{}"
# create an instance of SparkDataSample from a JSON string
spark_data_sample_instance = SparkDataSample.from_json(json)
# print the JSON string representation of the object
print(SparkDataSample.to_json())

# convert the object into a dict
spark_data_sample_dict = spark_data_sample_instance.to_dict()
# create an instance of SparkDataSample from a dict
spark_data_sample_from_dict = SparkDataSample.from_dict(spark_data_sample_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


