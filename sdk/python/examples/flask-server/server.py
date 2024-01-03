#!python3
from flask import Flask
import sys
import zrok
from zrok.model import ShareRequest
import atexit

app = Flask(__name__)
zrok_opts = {}
bindPort = 18081


@zrok.decor.zrok(opts=zrok_opts)
def runApp():
    from waitress import serve
    # the port is only used to integrate Zrok with frameworks that expect a "hostname:port" combo
    serve(app, port=bindPort)


@app.route('/')
def hello_world():
    print("received a request to /")
    return "Look! It's zrok!"


if __name__ == '__main__':
    root = zrok.environment.root.Load()
    try:
        shr = zrok.share.CreateShare(root=root, request=ShareRequest(
            BackendMode=zrok.model.TCP_TUNNEL_BACKEND_MODE,
            ShareMode=zrok.model.PUBLIC_SHARE_MODE,
            Frontends=['public'],
            Target="flask-server"
        ))
        shrToken = shr.Token
        print("Access server at the following endpoints: ", "\n".join(shr.FrontendEndpoints))

        def removeShare():
            zrok.share.DeleteShare(root=root, shr=shr)
            print("Deleted share")
        atexit.register(removeShare)
    except Exception as e:
        print("unable to create share", e)
        sys.exit(1)

    zrok_opts['cfg'] = zrok.decor.Opts(root=root, shrToken=shrToken, bindPort=bindPort)

    runApp()
