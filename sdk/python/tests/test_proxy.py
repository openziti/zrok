"""Tests for zrok2.proxy."""

import http.server
import socketserver
import threading
import urllib.parse

import pytest

from zrok2.model import Share
from zrok2.proxy import ProxyShare, _target_url


class ThreadingTCPServer(socketserver.ThreadingMixIn, socketserver.TCPServer):
    allow_reuse_address = True


def _serve(handler_cls):
    srv = ThreadingTCPServer(("127.0.0.1", 0), handler_cls)
    thread = threading.Thread(target=srv.serve_forever, daemon=True)
    thread.start()
    return srv


def _proxy_share(target):
    return ProxyShare(
        root=None,
        share=Share(Token="demo", FrontendEndpoints=[]),
        target=target,
    )


def test_target_url_rejects_absolute_paths():
    with pytest.raises(ValueError, match="absolute proxy paths"):
        _target_url("http://127.0.0.1:19191/base/", "http://127.0.0.1:19190/metadata")

    with pytest.raises(ValueError, match="absolute proxy paths"):
        _target_url("http://127.0.0.1:19191/base/", "//127.0.0.1:19190/metadata")


def test_proxy_rejects_encoded_absolute_url_path():
    class MetadataHandler(http.server.BaseHTTPRequestHandler):
        request_count = 0

        def do_GET(self):
            type(self).request_count += 1
            if self.path == "/metadata":
                self.send_response(200)
                self.end_headers()
                self.wfile.write(b"INTERNAL_METADATA_TOKEN")
                return
            self.send_response(404)
            self.end_headers()

        def log_message(self, fmt, *args):
            pass

    internal = _serve(MetadataHandler)
    try:
        target = "http://127.0.0.1:19191/base/"
        client = _proxy_share(target)._create_app().test_client()
        absolute_url = f"http://127.0.0.1:{internal.server_address[1]}/metadata"
        encoded = "/" + urllib.parse.quote(absolute_url, safe="")

        response = client.get(encoded)

        assert response.status_code == 400
        assert response.data != b"INTERNAL_METADATA_TOKEN"
        assert MetadataHandler.request_count == 0
    finally:
        internal.shutdown()
        internal.server_close()


def test_proxy_forwards_relative_paths_to_configured_target():
    class TargetHandler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            if self.path == "/base/resource":
                self.send_response(200)
                self.end_headers()
                self.wfile.write(b"target response")
                return
            self.send_response(404)
            self.end_headers()

        def log_message(self, fmt, *args):
            pass

    target = _serve(TargetHandler)
    try:
        target_url = f"http://127.0.0.1:{target.server_address[1]}/base/"
        client = _proxy_share(target_url)._create_app().test_client()

        response = client.get("/resource")

        assert response.status_code == 200
        assert response.data == b"target response"
    finally:
        target.shutdown()
        target.server_close()
