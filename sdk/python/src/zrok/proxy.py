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
from zrok.model import (PROXY_BACKEND_MODE, PUBLIC_SHARE_MODE, PRIVATE_SHARE_MODE, Share,
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
    verify_ssl: bool = True

    @classmethod
    def create(cls, root: Root, target: str, share_mode: str = PUBLIC_SHARE_MODE, unique_name: Optional[str] = None,
               frontends: Optional[List[str]] = None, verify_ssl: bool = True) -> 'ProxyShare':
        """
        Create a new proxy share, handling reservation and cleanup logic based on unique_name.

        Args:
            root: The zrok root environment
            target: Target URL or service to proxy to
            unique_name: Optional unique name for a reserved share
            frontends: Optional list of frontends to use, takes precedence over root's default_frontend
            verify_ssl: Whether to verify SSL certificates when forwarding requests.

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
                    unique_name=unique_name,
                    verify_ssl=verify_ssl
                )

        # Compose the share request
        share_frontends = []
        if share_mode == PUBLIC_SHARE_MODE:
            if frontends:
                share_frontends = frontends
            elif root.cfg and root.cfg.DefaultFrontend:
                share_frontends = [root.cfg.DefaultFrontend]
            else:
                share_frontends = ['public']

        share_request = ShareRequest(
            BackendMode=PROXY_BACKEND_MODE,
            ShareMode=share_mode,
            Target=target,
            Frontends=share_frontends,
            Reserved=True
        )
        if unique_name:
            share_request.UniqueName = unique_name

        # Create the share
        share = CreateShare(root=root, request=share_request)
        if share_mode == PUBLIC_SHARE_MODE:
            logger.debug(f"Created new proxy share with endpoints: {', '.join(share.FrontendEndpoints)}")
        elif share_mode == PRIVATE_SHARE_MODE:
            logger.debug(f"Created new private share with token: {share.Token}")

        # Create class instance and setup cleanup-at-exit if we reserved a random share token
        instance = cls(
            root=root,
            share=share,
            target=target,
            unique_name=unique_name,
            verify_ssl=verify_ssl
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
        """Register cleanup handler to release randomly generated shares on exit."""
        if not self._cleanup_registered:
            def cleanup():
                try:
                    ReleaseReservedShare(root=self.root, shr=self.share)
                    logger.info(f"Share {self.share.Token} released")
                except Exception as e:
                    logger.error(f"Error during cleanup: {e}")

            # Register for normal exit only
            atexit.register(cleanup)
            self._cleanup_registered = True
            return cleanup  # Return the cleanup function for reuse

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
                stream=True,
                verify=self.verify_ssl
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
        """Run the proxy server."""
        if not self._app:
            self._app = self._create_app()

        zrok_opts: Dict[str, Any] = {}
        zrok_opts['cfg'] = zrok.decor.Opts(root=self.root, shrToken=self.token, bindPort=DUMMY_PORT)

        @zrok.decor.zrok(opts=zrok_opts)
        def run_server():
            serve(self._app, port=DUMMY_PORT, _quiet=True)

        run_server()

    @property
    def endpoints(self) -> List[str]:
        """Get the frontend endpoints for this share."""
        return self.share.FrontendEndpoints

    @property
    def token(self) -> str:
        """Get the share token."""
        return self.share.Token
