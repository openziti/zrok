#!python3
from flask import Flask
import openziti
import sys
import zrok
from zrok.model import AccessRequest, ShareRequest
import atexit
import openziti

app = Flask(__name__)
zrok_opts = {}

@zrok.decor.zrok(opts=zrok_opts)
def runApp():
    from waitress import serve
    print("starting server through zrok")
    #the port is only used to integrate Zrok with frameworks that expect a "hostname:port" combo
    serve(app,port=18081)

@app.route('/')
def hello_world():  # put application's code here
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
        def removeShare():
            zrok.share.DeleteShare(root=root, shr=shr)
            print("Deleted share")
        atexit.register(removeShare)
    except Exception as e:
        print("unable to create share", e)
        sys.exit(1)

    zrok_opts['cfg'] = zrok.decor.Opts(root=root, shrToken=shrToken,bindPort=18081)
    
    try:
        acc = zrok.access.CreateAccess(root=root, request=AccessRequest(
            ShareToken=shrToken,
        ))
        def removeAccess():
            zrok.access.DeleteAccess(root=root, acc=acc)
            print("deleted access")
        atexit.register(removeAccess)
    except Exception as e:
        print("unable to create access", e)
        sys.exit(1)

    runApp()