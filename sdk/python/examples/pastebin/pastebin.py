#!python3
import argparse
import sys
import os
import zrok
import zrok.listener
import zrok.dialer
from zrok.model import AccessRequest, ShareRequest
import signal
import threading

exit_signal = threading.Event()

def signal_handler(signum, frame):
    print("\nCtrl-C detected. Next connection will close server")
    exit_signal.set()

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
        print("access your pastebin using 'pastebin.py pastefrom " + shr.Token + "'")

        try:
            with zrok.listener.Listener(shr.Token, root) as server:
                while not exit_signal.is_set():
                    conn, peer = server.accept()
                    with conn:
                        conn.sendall(data.encode('utf-8'))

        except KeyboardInterrupt:
            pass

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

    client = zrok.dialer.Dialer(options.shrToken, root)
    data = client.recv(1024)
    print(data.decode('utf-8'))
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
    signal.signal(signal.SIGINT, signal_handler)
    # Create a separate thread to run the server so we can respond to ctrl-c when in 'accept'
    server_thread = threading.Thread(target=options.func, args=[options])
    server_thread.start()

    server_thread.join()