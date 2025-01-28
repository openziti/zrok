import argparse
import atexit
import logging
import sys
import urllib.parse

import requests
from flask import Flask, Response, request
from waitress import serve
from zrok.model import ShareRequest, Share
from zrok.overview import EnvironmentAndResources, Overview

import zrok

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
target_url = None
zrok_opts = {}
bindPort = 18081

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


@app.route('/', defaults={'path': ''}, methods=['GET', 'POST', 'PUT', 'DELETE', 'HEAD', 'OPTIONS', 'PATCH'])
@app.route('/<path:path>', methods=['GET', 'POST', 'PUT', 'DELETE', 'HEAD', 'OPTIONS', 'PATCH'])
def proxy(path):
    global target_url
    logger.info(f"Incoming {request.method} request to {request.path}")
    logger.info(f"Headers: {dict(request.headers)}")

    # Forward the request to target URL
    full_url = urllib.parse.urljoin(target_url, request.path)
    logger.info(f"Forwarding to: {full_url}")

    # Copy request headers, excluding hop-by-hop headers
    headers = {k: v for k, v in request.headers.items() if k.lower() not in HOP_BY_HOP_HEADERS and k.lower() != 'host'}

    try:
        response = requests.request(
            method=request.method,
            url=full_url,
            headers=headers,
            data=request.get_data(),
            stream=True
        )

        logger.info(f"Response status: {response.status_code}")
        logger.info(f"Response headers: {dict(response.headers)}")

        # Filter out hop-by-hop headers from the response
        filtered_headers = {k: v for k, v in response.headers.items() if k.lower() not in HOP_BY_HOP_HEADERS}

        return Response(
            response.iter_content(chunk_size=8192),
            status=response.status_code,
            headers=filtered_headers
        )

    except Exception as e:
        logger.error(f"Proxy error: {str(e)}", exc_info=True)
        return str(e), 502


@zrok.decor.zrok(opts=zrok_opts)
def run_proxy():
    # the port is only used to integrate zrok with frameworks that expect a "hostname:port" combo
    serve(app, port=bindPort)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Start a zrok proxy server')
    parser.add_argument('target_url', help='Target URL to proxy requests to')
    parser.add_argument('-n', '--unique-name', help='Unique name for the proxy instance')
    args = parser.parse_args()

    target_url = args.target_url
    logger.info("=== Starting proxy server ===")
    logger.info(f"Target URL: {target_url}")
    logger.info(f"Logging level: {logger.getEffectiveLevel()}")

    root = zrok.environment.root.Load()
    my_env = EnvironmentAndResources(
        environment=None,
        shares=[]
    )
    overview = Overview.create(root=root)
    for env_stuff in overview.environments:
        if env_stuff.environment.z_id == root.env.ZitiIdentity:
            my_env = EnvironmentAndResources(
                environment=env_stuff.environment,
                shares=env_stuff.shares
            )
            break

    if my_env:
        logger.debug(
            f"Found environment in overview with Ziti identity "
            f"matching local environment: {my_env.environment.z_id}"
        )
    else:
        logger.error("No matching environment found")
        sys.exit(1)

    existing_reserved_share = None
    for share in my_env.shares:
        if share.token == args.unique_name:
            existing_reserved_share = share
            break

    if existing_reserved_share:
        logger.debug(f"Found existing share with token: {existing_reserved_share.token}")
        shr = Share(Token=existing_reserved_share.token, FrontendEndpoints=[existing_reserved_share.frontend_endpoint])
    else:
        logger.debug(f"No existing share found with token: {args.unique_name}")
        share_request = ShareRequest(
            BackendMode=zrok.model.PROXY_BACKEND_MODE,
            ShareMode=zrok.model.PUBLIC_SHARE_MODE,
            Frontends=['public'],
            Target="http-proxy",
            Reserved=True
        )
        if args.unique_name:
            share_request.UniqueName = args.unique_name

        shr = zrok.share.CreateShare(root=root, request=share_request)

    def cleanup():
        zrok.share.ReleaseReservedShare(root=root, shr=shr)
        logger.info(f"Share {shr.Token} released")
    if not args.unique_name:
        atexit.register(cleanup)

    zrok_opts['cfg'] = zrok.decor.Opts(root=root, shrToken=shr.Token, bindPort=bindPort)

    logger.info(f"Access proxy at: {', '.join(shr.FrontendEndpoints)}")

    run_proxy()
