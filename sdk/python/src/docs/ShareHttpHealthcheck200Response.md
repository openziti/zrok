# ShareHttpHealthcheck200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**healthy** | **bool** |  | [optional] 
**error** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.share_http_healthcheck200_response import ShareHttpHealthcheck200Response

# TODO update the JSON string below
json = "{}"
# create an instance of ShareHttpHealthcheck200Response from a JSON string
share_http_healthcheck200_response_instance = ShareHttpHealthcheck200Response.from_json(json)
# print the JSON string representation of the object
print(ShareHttpHealthcheck200Response.to_json())

# convert the object into a dict
share_http_healthcheck200_response_dict = share_http_healthcheck200_response_instance.to_dict()
# create an instance of ShareHttpHealthcheck200Response from a dict
share_http_healthcheck200_response_from_dict = ShareHttpHealthcheck200Response.from_dict(share_http_healthcheck200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


