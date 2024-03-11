from zrok.environment.root import Root
from zrok_api.models import ShareRequest, UnshareRequest, AuthUser
from zrok_api.api import ShareApi
from zrok import model


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
        res = ShareApi(zrok).share(body=out)
    except Exception as e:
        raise Exception("unable to create share", e)

    return model.Share(Token=res.shr_token,
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
                         shr_token=shr.Token)

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        ShareApi(zrok).unshare(body=req)
    except Exception as e:
        raise Exception("error deleting share", e)
