"""
Proxy share management functionality for the zrok SDK.
"""

import atexit
import logging
import urllib.parse
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

import requests
from flask import Flask, Response, request
from waitress import serve
from zrok.environment.root import Root
from zrok.model import (PROXY_BACKEND_MODE, PUBLIC_SHARE_MODE, Share,
                        ShareRequest)
from zrok.overview import Overview
from zrok.share import CreateShare, ReleaseReservedShare

import zrok

logger = logging.getLogger(__name__)

# List of hop-by-hop headers that should not be returned to the viewer
HOP_BY_HOP_HEADERS = {
    'connection',
    'keep-alive',
    'proxy-authenticate',
    'proxy-authorization',
    'te',
    'trailers',
    'transfer-encoding',
    'upgrade'
}

# The proxy only listens on the zrok socket, the port used to initialize the Waitress server is not actually bound or
# listening
DUMMY_PORT = 18081


@dataclass
class ProxyShare:
    """Represents a proxy share with its configuration and state."""
    root: Root
    share: Share
    target: str
    unique_name: Optional[str] = None
    _cleanup_registered: bool = False
    _app: Optional[Flask] = None

    @classmethod
    def create(cls, root: Root, target: str, unique_name: Optional[str] = None) -> 'ProxyShare':
        """
        Create a new proxy share, handling reservation and cleanup logic based on unique_name.

        Args:
            root: The zrok root environment
            target: Target URL or service to proxy to
            unique_name: Optional unique name for a reserved share

        Returns:
            ProxyShare instance configured with the created share
        """
        # First check if we have an existing reserved share with this name
        if unique_name:
            existing_share = cls._find_existing_share(root, unique_name)
            if existing_share:
                logger.debug(f"Found existing share with token: {existing_share.Token}")
                return cls(
                    root=root,
                    share=existing_share,
                    target=target,
                    unique_name=unique_name
                )

        # Create new share request
        share_request = ShareRequest(
            BackendMode=PROXY_BACKEND_MODE,
            ShareMode=PUBLIC_SHARE_MODE,
            Target="http-proxy",
            Frontends=['public'],
            Reserved=bool(unique_name)
        )
        if unique_name:
            share_request.UniqueName = unique_name

        # Create the share
        share = CreateShare(root=root, request=share_request)
        logger.info(f"Created new proxy share with endpoints: {', '.join(share.FrontendEndpoints)}")

        # Create instance and setup cleanup if needed
        instance = cls(
            root=root,
            share=share,
            target=target,
            unique_name=unique_name
        )
        if not unique_name:
            instance.register_cleanup()
        return instance

    @staticmethod
    def _find_existing_share(root: Root, unique_name: str) -> Optional[Share]:
        """Find an existing share with the given unique name."""
        overview = Overview.create(root=root)
        for env in overview.environments:
            if env.environment.z_id == root.env.ZitiIdentity:
                for share in env.shares:
                    if share.token == unique_name:
                        return Share(Token=share.token, FrontendEndpoints=[share.frontend_endpoint])
        return None

    def register_cleanup(self):
        """Register cleanup handler to release the share on exit."""
        if not self._cleanup_registered:
            def cleanup():
                ReleaseReservedShare(root=self.root, shr=self.share)
                logger.info(f"Share {self.share.Token} released")
            atexit.register(cleanup)
            self._cleanup_registered = True

    def _create_app(self) -> Flask:
        """Create and configure the Flask app for proxying."""
        app = Flask(__name__)

        @app.route('/', defaults={'path': ''}, methods=['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'])
        @app.route('/<path:path>', methods=['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'])
        def proxy(path):
            # Construct the target URL
            url = urllib.parse.urljoin(self.target, path)

            # Forward the request
            resp = requests.request(
                method=request.method,
                url=url,
                headers={key: value for (key, value) in request.headers
                         if key.lower() not in HOP_BY_HOP_HEADERS},
                data=request.get_data(),
                cookies=request.cookies,
                allow_redirects=False,
                stream=True
            )

            # Create the response
            excluded_headers = HOP_BY_HOP_HEADERS.union({'host'})
            headers = [(name, value) for (name, value) in resp.raw.headers.items()
                       if name.lower() not in excluded_headers]

            return Response(
                resp.iter_content(chunk_size=10*1024),
                status=resp.status_code,
                headers=headers
            )
        return app

    def run(self):
        """Start the proxy server."""
        if self._app is None:
            self._app = self._create_app()

        # Create options dictionary for zrok decorator
        zrok_opts: Dict[str, Any] = {}
        zrok_opts['cfg'] = zrok.decor.Opts(root=self.root, shrToken=self.token, bindPort=DUMMY_PORT)

        @zrok.decor.zrok(opts=zrok_opts)
        def run_server():
            serve(self._app, port=DUMMY_PORT)
        run_server()

    @property
    def endpoints(self) -> List[str]:
        """Get the frontend endpoints for this share."""
        return self.share.FrontendEndpoints

    @property
    def token(self) -> str:
        """Get the share token."""
        return self.share.Token
