import openziti
from zrok.environment.root import Root

class MonkeyPatch(openziti.monkeypatch):
    def __init__(self, bindHost: str, bindPort: int, root: Root, shrToken: str, **kwargs):
        zif = root.ZitiIdentityNamed(root.EnvironmentIdentityName())
        cfg = dict(ztx=openziti.load(zif), service=shrToken)
        super(MonkeyPatch, self).__init__(bindings={(bindHost, bindPort):cfg})

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        super(MonkeyPatch, self).__exit__(exc_type, exc_val, exc_tb)

def zrok(bindHost: str, bindPort: int, root: Root, shrToken: str, **zkwargs):
    def zrockify_func(func):
        def zrockified(*args, **kwargs):
            with MonkeyPatch(bindHost=bindHost, bindPort=bindPort, root=root, shrToken=shrToken, **zkwargs):
                func(*args, **kwargs)
        return zrockified
    return zrockify_func