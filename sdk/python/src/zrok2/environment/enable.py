import socket

from zrok_api.api import EnvironmentApi
from zrok_api.models.enable_request import EnableRequest as ApiEnableRequest
from zrok_api.models.disable_request import DisableRequest as ApiDisableRequest
from zrok_api.exceptions import ApiException
from zrok2.environment.root import Root, Environment


def enable(root: Root, token: str, description: str = None, host: str = None) -> Environment:
    """Enable a zrok environment from an account token.

    Analogous to ``zrok2 enable <token>``. Safe to call when already enabled
    (returns existing environment without modification).

    Args:
        root: The Root environment context (from ``root.Load()``).
        token: Account enable token.
        description: Optional description for this environment.
        host: Optional hostname override (defaults to system hostname).

    Returns:
        The Environment dataclass populated with token, identity, and endpoint.
    """
    if root.IsEnabled():
        return root.env

    if host is None:
        host = socket.gethostname()

    if description is None:
        description = ""

    # Temporarily set the account token so Client() can authenticate
    api_endpoint = root.ApiEndpoint()
    root.env.Token = token
    root.env.ApiEndpoint = api_endpoint.endpoint

    try:
        zrok = root.Client()
    except Exception as e:
        root.env = Environment()
        raise Exception("error getting zrok client", e)

    try:
        req = ApiEnableRequest(description=description, host=host)
        env_api = EnvironmentApi(zrok)
        custom_headers = {
            'Accept': 'application/json, application/zrok.v1+json'
        }
        resp = env_api.enable_with_http_info(body=req, _headers=custom_headers)

        if hasattr(resp, 'data') and resp.data is not None:
            res = resp.data
        else:
            raise Exception("invalid response from enable API")
    except ApiException as e:
        root.env = Environment()
        if "Unsupported content type: application/zrok.v1+json" in str(e) and hasattr(e, 'body'):
            import json
            try:
                res_dict = json.loads(e.body)

                class EnableResponse:
                    def __init__(self, identity, cfg):
                        self.identity = identity
                        self.cfg = cfg

                res = EnableResponse(
                    identity=res_dict.get('identity', ''),
                    cfg=res_dict.get('cfg', '')
                )
            except (json.JSONDecodeError, ValueError, AttributeError) as json_err:
                raise Exception(f"unable to parse enable response: {str(json_err)}") from e
        else:
            raise Exception("unable to enable environment", e)
    except Exception as e:
        root.env = Environment()
        raise Exception("unable to enable environment", e)

    env = Environment(
        Token=token,
        ZitiIdentity=res.identity,
        ApiEndpoint=api_endpoint.endpoint,
    )

    root.SetEnvironment(env)
    root.SaveZitiIdentityNamed(root.EnvironmentIdentityName(), res.cfg)

    return env


def disable(root: Root) -> None:
    """Disable the current zrok environment.

    Analogous to ``zrok2 disable``. No-op if not currently enabled.

    Args:
        root: The Root environment context.
    """
    if not root.IsEnabled():
        return

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        req = ApiDisableRequest(identity=root.env.ZitiIdentity)
        env_api = EnvironmentApi(zrok)
        custom_headers = {
            'Accept': 'application/json, application/zrok.v1+json'
        }
        env_api.disable_with_http_info(body=req, _headers=custom_headers)
    except ApiException as e:
        if "Unsupported content type: application/zrok.v1+json" in str(e) and (200 <= e.status <= 299):
            pass
        else:
            raise Exception("unable to disable environment", e)
    except Exception as e:
        raise Exception("unable to disable environment", e)

    root.DeleteEnvironment()
