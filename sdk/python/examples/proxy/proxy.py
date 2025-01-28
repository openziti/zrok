#!/usr/bin/env python3

"""
Example of using zrok's proxy facility to create an HTTP proxy server.

This example demonstrates how to:
1. Create a proxy share (optionally with a unique name for persistence)
2. Handle HTTP requests/responses through the proxy
3. Automatically clean up non-reserved shares on exit
"""

import argparse
import logging

import zrok
from zrok.proxy import ProxyShare

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Start a zrok proxy server')
    parser.add_argument('target_url', help='Target URL to proxy requests to')
    parser.add_argument('-n', '--unique-name', help='Unique name for the proxy instance')
    args = parser.parse_args()

    logger.info("=== Starting proxy server ===")
    logger.info(f"Target URL: {args.target_url}")

    # Load environment and create proxy share
    root = zrok.environment.root.Load()
    proxy_share = ProxyShare.create(
        root=root,
        target=args.target_url,
        unique_name=args.unique_name
    )
    
    # Log access information and start the proxy
    logger.info(f"Access proxy at: {', '.join(proxy_share.endpoints)}")
    proxy_share.run()


if __name__ == '__main__':
    main()
