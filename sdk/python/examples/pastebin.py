#!python3
import argparse
import sys
import os
import zrok
from zrok.model import AccessRequest, ShareRequest
from http.server import BaseHTTPRequestHandler, HTTPServer
import urllib3

class MyServer(BaseHTTPRequestHandler):
    def __init__(self, data, *args, **kwargs):
        self.data = data
        super(MyServer, self).__init__(*args, **kwargs)

    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.send_header("Content-length", len(self.data))
        self.end_headers()
        self.wfile.write(bytes(self.data, "utf-8"))

class copyto:
    def handle(self, *args, **kwargs):
        root = zrok.environment.root.Load()
        
        try:
            shr = zrok.share.CreateShare(root=root, request=ShareRequest(
                BackendMode=zrok.model.TCP_TUNNEL_BACKEND_MODE,
                ShareMode=zrok.model.PRIVATE_SHARE_MODE,
                Target="pastebin"
            ))
        except Exception as e:
            print("unable to create share", e)
            sys.exit(1)

        data = self.loadData()
        def handler(*args):
            MyServer(data, *args)
        zrok.monkeypatch(bindHost="127.0.0.1", bindPort=8082, root=root, shrToken=shr.Token)
        webServer = HTTPServer(("127.0.0.1", 8082), handler)
        print("access your pastebin using 'pastebin.py pastefrom " + shr.Token + "'")

        try:
            webServer.serve_forever(poll_interval=600)
        except KeyboardInterrupt:
            pass

        webServer.server_close()
        zrok.share.DeleteShare(root, shr)
        print("Server stopped.")
        

    def loadData(self):
        if not os.isatty(sys.stdin.fileno()):
            return sys.stdin.read()
        else:
            raise Exception("'copyto' requires input from stdin; direct your paste buffer into stdin")

def pastefrom(options):
    root = zrok.environment.root.Load()

    try:
        acc = zrok.access.CreateAccess(root=root, request=AccessRequest(
            ShareToken=options.shrToken,
        ))
    except Exception as e:
        print("unable to create access", e)
        sys.exit(1)

    zrok.monkeypatch(bindHost="127.0.0.1", bindPort=8082, root=root, shrToken=options.shrToken)

    http = urllib3.PoolManager()
    try:
        r = http.request('GET', "http://" + options.shrToken)
    except Exception as e:
        print("Error on request: ", e)
        zrok.access.DeleteAccess(root, acc)
        return
    print(r.data.decode('utf-8'))
    try:
        zrok.access.DeleteAccess(root, acc)
    except Exception as e:
        print("unable to delete access", e)
        sys.exit(1)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers()
    subparsers.required = True

    c = copyto()
    parser_copyto = subparsers.add_parser('copyto')
    parser_copyto.set_defaults(func=c.handle)

    parser_pastefrom = subparsers.add_parser('pastefrom')
    parser_pastefrom.set_defaults(func=pastefrom)
    parser_pastefrom.add_argument("shrToken")

    options = parser.parse_args()
    options.func(options)