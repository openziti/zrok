# ApplyLimitClassesRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **str** |  | [optional] 
**limit_class_ids** | **List[int]** |  | [optional] 

## Example

```python
from zrok_api.models.apply_limit_classes_request import ApplyLimitClassesRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ApplyLimitClassesRequest from a JSON string
apply_limit_classes_request_instance = ApplyLimitClassesRequest.from_json(json)
# print the JSON string representation of the object
print(ApplyLimitClassesRequest.to_json())

# convert the object into a dict
apply_limit_classes_request_dict = apply_limit_classes_request_instance.to_dict()
# create an instance of ApplyLimitClassesRequest from a dict
apply_limit_classes_request_from_dict = ApplyLimitClassesRequest.from_dict(apply_limit_classes_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


