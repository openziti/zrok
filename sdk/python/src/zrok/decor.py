from dataclasses import dataclass
import openziti
from zrok.environment.root import Root


@dataclass
class Opts:
    root: Root
    shrToken: str
    bindPort: int
    bindHost: str = ""


class MonkeyPatch(openziti.monkeypatch):
    def __init__(self, opts: {}, *args, **kwargs):
        zif = opts['cfg'].root.ZitiIdentityNamed(opts['cfg'].root.EnvironmentIdentityName())
        cfg = dict(ztx=openziti.load(zif), service=opts['cfg'].shrToken)
        super(MonkeyPatch, self).__init__(bindings={(opts['cfg'].bindHost, opts['cfg'].bindPort): cfg}, *args, **kwargs)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        super(MonkeyPatch, self).__exit__(exc_type, exc_val, exc_tb)


def zrok(opts: {}, *zargs, **zkwargs):
    def zrockify_func(func):
        def zrockified(*args, **kwargs):
            with MonkeyPatch(opts=opts, *zargs, **zkwargs):
                func(*args, **kwargs)
        return zrockified
    return zrockify_func
