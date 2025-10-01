import json
from dataclasses import dataclass, field
from typing import List

import urllib3
from zrok.environment.root import Root
from zrok_api.models.environment import Environment
from zrok_api.models.environment_and_resources import EnvironmentAndResources
from zrok_api.models.share import Share


@dataclass
class Overview:
    environments: List[EnvironmentAndResources] = field(default_factory=list)

    @classmethod
    def create(cls, root: Root) -> 'Overview':
        if not root.IsEnabled():
            raise Exception("environment is not enabled; enable with 'zrok enable' first!")

        http = urllib3.PoolManager()
        apiEndpoint = root.ApiEndpoint().endpoint
        try:
            response = http.request(
                'GET',
                apiEndpoint + "/api/v2/overview",
                headers={
                    "X-TOKEN": root.env.Token
                })
        except Exception as e:
            raise Exception("unable to get account overview", e)

        json_data = json.loads(response.data.decode('utf-8'))
        overview = cls()

        for env_data in json_data.get('environments', []):
            env_dict = env_data['environment']
            # Map the JSON keys to the Environment class parameters
            environment = Environment(
                description=env_dict.get('description'),
                host=env_dict.get('host'),
                address=env_dict.get('address'),
                z_id=env_dict.get('zId'),
                activity=env_dict.get('activity'),
                limited=env_dict.get('limited'),
                created_at=env_dict.get('createdAt'),
                updated_at=env_dict.get('updatedAt')
            )

            # Create Shares object from share data
            share_list = []
            for share_data in env_data.get('shares', []):
                share = Share(
                    share_token=share_data.get('shareToken'),
                    z_id=share_data.get('zId'),
                    share_mode=share_data.get('shareMode'),
                    backend_mode=share_data.get('backendMode'),
                    frontend_selection=share_data.get('frontendSelection'),
                    frontend_endpoint=share_data.get('frontendEndpoint'),
                    backend_proxy_endpoint=share_data.get('backendProxyEndpoint'),
                    reserved=share_data.get('reserved'),
                    activity=share_data.get('activity'),
                    limited=share_data.get('limited'),
                    created_at=share_data.get('createdAt'),
                    updated_at=share_data.get('updatedAt')
                )
                share_list.append(share)

            # Create EnvironmentAndResources object
            env_resources = EnvironmentAndResources(
                environment=environment,
                shares=share_list,
                frontends=[]
            )
            overview.environments.append(env_resources)

        return overview
