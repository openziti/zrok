# ShareHttpHealthcheckRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 
**share_token** | **str** |  | [optional] 
**http_verb** | **str** |  | [optional] 
**endpoint** | **str** |  | [optional] 
**expected_http_response** | **float** |  | [optional] 
**timeout_ms** | **float** |  | [optional] 

## Example

```python
from zrok_api.models.share_http_healthcheck_request import ShareHttpHealthcheckRequest

# TODO update the JSON string below
json = "{}"
# create an instance of ShareHttpHealthcheckRequest from a JSON string
share_http_healthcheck_request_instance = ShareHttpHealthcheckRequest.from_json(json)
# print the JSON string representation of the object
print(ShareHttpHealthcheckRequest.to_json())

# convert the object into a dict
share_http_healthcheck_request_dict = share_http_healthcheck_request_instance.to_dict()
# create an instance of ShareHttpHealthcheckRequest from a dict
share_http_healthcheck_request_from_dict = ShareHttpHealthcheckRequest.from_dict(share_http_healthcheck_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


