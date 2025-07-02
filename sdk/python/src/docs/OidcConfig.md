# OidcConfig


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**provider_id** | **str** |  | [optional] 
**issuer_url** | **str** |  | [optional] 
**authz_url_params** | **List[str]** |  | [optional] 
**cookie_domain** | **str** |  | [optional] 
**client_id** | **str** |  | [optional] 
**client_secret** | **str** |  | [optional] 
**scopes** | **List[str]** |  | [optional] 
**max_session_duration** | **str** |  | [optional] 
**idle_session_duration** | **str** |  | [optional] 
**userinfo_refresh_interval** | **str** |  | [optional] 

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


