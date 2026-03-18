#!python3
from flask import Flask
import sys
import os
import zrok2
from zrok2.model import ShareRequest
import atexit

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from common import get_root  # noqa: E402

app = Flask(__name__)
zrok_opts = {}
bindPort = 18081


@zrok2.decor.zrok(opts=zrok_opts)
def runApp():
    from waitress import serve
    # the port is only used to integrate zrok with frameworks that expect a "hostname:port" combo
    serve(app, port=bindPort)


@app.route('/')
def hello_world():
    print("received a request to /")
    return "Look! It's zrok!"


if __name__ == '__main__':
    root = get_root()
    try:
        shr = zrok2.share.CreateShare(root=root, request=ShareRequest(
            BackendMode=zrok2.model.PROXY_BACKEND_MODE,
            ShareMode=zrok2.model.PUBLIC_SHARE_MODE,
            Frontends=['public'],
            Target="http-server"
        ))
        shrToken = shr.Token
        print("Access server at the following endpoints: ", "\n".join(shr.FrontendEndpoints or []))

        def removeShare():
            zrok2.share.DeleteShare(root=root, shr=shr)
            print("Deleted share")
        atexit.register(removeShare)
    except Exception as e:
        print("unable to create share", e)
        sys.exit(1)

    zrok_opts['cfg'] = zrok2.decor.Opts(root=root, shrToken=shrToken, bindPort=bindPort)

    runApp()
