# AgentStatusRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**env_zid** | **str** |  | [optional] 

## Example

```python
from zrok_api.models.agent_status_request import AgentStatusRequest

# TODO update the JSON string below
json = "{}"
# create an instance of AgentStatusRequest from a JSON string
agent_status_request_instance = AgentStatusRequest.from_json(json)
# print the JSON string representation of the object
print(AgentStatusRequest.to_json())

# convert the object into a dict
agent_status_request_dict = agent_status_request_instance.to_dict()
# create an instance of AgentStatusRequest from a dict
agent_status_request_from_dict = AgentStatusRequest.from_dict(agent_status_request_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


