from zrok_api.api import ShareApi
from zrok.environment.root import Root
from zrok_api.models.auth_user import AuthUser
from zrok_api.models.share_request import ShareRequest
from zrok_api.models.unshare_request import UnshareRequest
import json
from zrok_api.exceptions import ApiException
import zrok.model as model


class Share():
    root: Root
    request: model.ShareRequest
    share: model.Share

    def __init__(self, root: Root, request: model.ShareRequest):
        self.root = root
        self.request = request

    def __enter__(self) -> model.Share:
        self.share = CreateShare(root=self.root, request=self.request)
        return self.share

    def __exit__(self, exception_type, exception_value, exception_traceback):
        if not self.request.Reserved:
            DeleteShare(root=self.root, shr=self.share)


def CreateShare(root: Root, request: model.ShareRequest) -> model.Share:
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    match request.ShareMode:
        case model.PRIVATE_SHARE_MODE:
            out = __newPrivateShare(root, request)
        case model.PUBLIC_SHARE_MODE:
            out = __newPublicShare(root, request)
        case _:
            raise Exception("unknown share mode " + request.ShareMode)
    out.reserved = request.Reserved
    if request.Reserved:
        out.unique_name = request.UniqueName

    if len(request.BasicAuth) > 0:
        out.auth_scheme = model.AUTH_SCHEME_BASIC
        for pair in request.BasicAuth:
            tokens = pair.split(":")
            if len(tokens) == 2:
                out.auth_users.append(AuthUser(username=tokens[0].strip(), password=tokens[1].strip()))
            else:
                raise Exception("invalid username:password pair: " + pair)

    if request.OauthProvider != "":
        out.auth_scheme = model.AUTH_SCHEME_OAUTH

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        # Use share_with_http_info to get access to the HTTP info and handle custom response format
        share_api = ShareApi(zrok)
        # Add Accept header to handle the custom content type
        custom_headers = {
            'Accept': 'application/json, application/zrok.v1+json'
        }

        response_data = share_api.share_with_http_info(
            body=out,
            _headers=custom_headers
        )

        # Parse response
        if hasattr(response_data, 'data') and response_data.data is not None:
            res = response_data.data
        else:
            raise Exception("invalid response from server")

    except ApiException as e:
        # If it's a content type error, try to parse the raw JSON
        if "Unsupported content type: application/zrok.v1+json" in str(e) and hasattr(e, 'body'):
            try:
                # Parse the response body directly
                res_dict = json.loads(e.body)
                # Create a response object with the expected fields

                class ShareResponse:
                    def __init__(self, share_token, frontend_proxy_endpoints):
                        self.share_token = share_token
                        self.frontend_proxy_endpoints = frontend_proxy_endpoints

                res = ShareResponse(
                    share_token=res_dict.get('shareToken', ''),
                    frontend_proxy_endpoints=res_dict.get('frontendProxyEndpoints', [])
                )
            except (json.JSONDecodeError, ValueError, AttributeError) as json_err:
                raise Exception(f"unable to parse API response: {str(json_err)}") from e
        else:
            raise Exception("unable to create share", e)
    except Exception as e:
        raise Exception("unable to create share", e)

    return model.Share(Token=res.share_token,
                       FrontendEndpoints=res.frontend_proxy_endpoints)


def __newPrivateShare(root: Root, request: model.ShareRequest) -> ShareRequest:
    return ShareRequest(env_zid=root.env.ZitiIdentity,
                        share_mode=request.ShareMode,
                        backend_mode=request.BackendMode,
                        backend_proxy_endpoint=request.Target,
                        auth_scheme=model.AUTH_SCHEME_NONE,
                        permission_mode=request.PermissionMode,
                        access_grants=request.AccessGrants
                        )


def __newPublicShare(root: Root, request: model.ShareRequest) -> ShareRequest:
    ret = ShareRequest(env_zid=root.env.ZitiIdentity,
                       share_mode=request.ShareMode,
                       frontend_selection=request.Frontends,
                       backend_mode=request.BackendMode,
                       backend_proxy_endpoint=request.Target,
                       auth_scheme=model.AUTH_SCHEME_NONE,
                       oauth_email_domains=request.OauthEmailAddressPatterns,
                       oauth_authorization_check_interval=request.OauthAuthorizationCheckInterval,
                       permission_mode=request.PermissionMode,
                       access_grants=request.AccessGrants
                       )
    if request.OauthProvider != "":
        ret.oauth_provider = request.OauthProvider

    return ret


def DeleteShare(root: Root, shr: model.Share):
    req = UnshareRequest(env_zid=root.env.ZitiIdentity,
                         share_token=shr.Token)

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        # Add Accept header to handle the custom content type
        share_api = ShareApi(zrok)
        custom_headers = {
            'Accept': 'application/json, application/zrok.v1+json'
        }

        # Use unshare_with_http_info to get access to the HTTP info
        share_api.unshare_with_http_info(
            body=req,
            _headers=custom_headers
        )
    except ApiException as e:
        # If it's a content type error but the operation was likely successful, don't propagate the error
        if "Unsupported content type: application/zrok.v1+json" in str(e) and (200 <= e.status <= 299):
            # The operation was likely successful despite the content type error
            pass
        else:
            raise Exception("error deleting share", e)
    except Exception as e:
        raise Exception("error deleting share", e)


def ReleaseReservedShare(root: Root, shr: model.Share):
    req = UnshareRequest(env_zid=root.env.ZitiIdentity,
                         share_token=shr.Token,
                         reserved=True)

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        # Add Accept header to handle the custom content type
        share_api = ShareApi(zrok)
        custom_headers = {
            'Accept': 'application/json, application/zrok.v1+json'
        }

        # Use unshare_with_http_info to get access to the HTTP info
        share_api.unshare_with_http_info(
            body=req,
            _headers=custom_headers
        )
    except ApiException as e:
        # If it's a content type error but the operation was likely successful, don't propagate the error
        if "Unsupported content type: application/zrok.v1+json" in str(e) and (200 <= e.status <= 299):
            # The operation was likely successful despite the content type error
            pass
        else:
            raise Exception("error releasing share", e)
    except Exception as e:
        raise Exception("error releasing share", e)
