#!python3
import argparse
import atexit
import sys
import os
import zrok2
import zrok2.listener
import zrok2.dialer
from zrok2.model import AccessRequest, ShareRequest
import signal
import threading

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from common import get_root  # noqa: E402

exit_signal = threading.Event()


def signal_handler(signum, frame):
    print("\nCtrl-C detected. Next connection will close server")
    exit_signal.set()


class copyto:
    def handle(self, *args, **kwargs):
        root = get_root()

        try:
            shr = zrok2.share.CreateShare(root=root, request=ShareRequest(
                BackendMode=zrok2.model.TCP_TUNNEL_BACKEND_MODE,
                ShareMode=zrok2.model.PRIVATE_SHARE_MODE,
                Target="pastebin"
            ))
        except Exception as e:
            print("unable to create share", e)
            sys.exit(1)

        def removeShare():
            try:
                zrok2.share.DeleteShare(root, shr)
            except Exception as e:
                print("unable to delete share", e)
                sys.exit(1)
        atexit.register(removeShare)

        data = self.loadData()
        print("access your pastebin using 'pastebin.py pastefrom " + shr.Token + "'")

        with zrok2.listener.Listener(shr.Token, root) as server:
            while not exit_signal.is_set():
                conn, peer = server.accept()
                with conn:
                    conn.sendall(data.encode('utf-8'))

        print("Server stopped.")

    def loadData(self):
        if not os.isatty(sys.stdin.fileno()):
            return sys.stdin.read()
        else:
            raise Exception("'copyto' requires input from stdin; direct your paste buffer into stdin")


def pastefrom(options):
    root = get_root()

    try:
        acc = zrok2.access.CreateAccess(root=root, request=AccessRequest(
            ShareToken=options.shrToken,
        ))
    except Exception as e:
        print("unable to create access", e)
        sys.exit(1)

    def removeAccess():
        try:
            zrok2.access.DeleteAccess(root, acc)
        except Exception as e:
            print("unable to delete access", e)
            sys.exit(1)
    atexit.register(removeAccess)

    client = zrok2.dialer.Dialer(options.shrToken, root)
    data = client.recv(1024)
    print(data.decode('utf-8'))


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
