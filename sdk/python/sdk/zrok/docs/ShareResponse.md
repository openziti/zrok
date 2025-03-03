# ShareResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**frontend_proxy_endpoints** | **List[str]** |  | [optional] 
**share_token** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.share_response import ShareResponse

# TODO update the JSON string below
json = "{}"
# create an instance of ShareResponse from a JSON string
share_response_instance = ShareResponse.from_json(json)
# print the JSON string representation of the object
print(ShareResponse.to_json())

# convert the object into a dict
share_response_dict = share_response_instance.to_dict()
# create an instance of ShareResponse from a dict
share_response_from_dict = ShareResponse.from_dict(share_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


