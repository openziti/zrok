from zrok.environment.root import Root
from zrok_api.models import AccessRequest, UnaccessRequest
from zrok_api.api import ShareApi
from zrok import model

def CreateAccess(root: Root, request: model.AccessRequest) -> model.Access:
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    out = AccessRequest(shr_token=request.ShareToken,
                        env_zid=root.env.ZitiIdentity)

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)
    try:
        res = ShareApi(zrok).access(body=out)
    except Exception as e:
        raise Exception("unable to create access", e)
    return model.Access(Token=res.frontend_token,
                       ShareToken=request.ShareToken,
                       BackendMode=res.backend_mode)

def DeleteAccess(root: Root, acc: model.Access):
    req = UnaccessRequest(frontend_token=acc.Token,
                          shr_token=acc.ShareToken,
                          env_zid=root.env.ZitiIdentity)
    
    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)
    
    try:
        ShareApi(zrok).unaccess(body=req)
    except Exception as e:
        raise Exception("error deleting access", e)