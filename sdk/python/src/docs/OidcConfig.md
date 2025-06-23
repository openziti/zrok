# OidcConfig


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**provider_name** | **str** |  | [optional] 
**client_id** | **str** |  | [optional] 
**scopes** | **List[str]** |  | [optional] 
**auth_url** | **str** |  | [optional] 
**token_url** | **str** |  | [optional] 
**email_endpoint** | **str** |  | [optional] 
**email_path** | **str** |  | [optional] 
**supports_pkce** | **bool** |  | [optional] 
**allowed_email_filters** | **List[str]** |  | [optional] 
**auth_timeout** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.oidc_config import OidcConfig

# TODO update the JSON string below
json = "{}"
# create an instance of OidcConfig from a JSON string
oidc_config_instance = OidcConfig.from_json(json)
# print the JSON string representation of the object
print(OidcConfig.to_json())

# convert the object into a dict
oidc_config_dict = oidc_config_instance.to_dict()
# create an instance of OidcConfig from a dict
oidc_config_from_dict = OidcConfig.from_dict(oidc_config_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


