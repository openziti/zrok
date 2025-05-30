# AgentStatus200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**version** | **str** |  | [optional] 
**shares** | [**List[Share]**](Share.md) |  | [optional] 

## Example

```python
from zrok_api.models.agent_status200_response import AgentStatus200Response

# TODO update the JSON string below
json = "{}"
# create an instance of AgentStatus200Response from a JSON string
agent_status200_response_instance = AgentStatus200Response.from_json(json)
# print the JSON string representation of the object
print(AgentStatus200Response.to_json())

# convert the object into a dict
agent_status200_response_dict = agent_status200_response_instance.to_dict()
# create an instance of AgentStatus200Response from a dict
agent_status200_response_from_dict = AgentStatus200Response.from_dict(agent_status200_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


