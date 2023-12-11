from setuptools import setup, find_packages  # noqa: H301
import os

NAME = "zrok_sdk"
VERSION = "0.4.0.dev"
try:
    VERSION = os.environ['ZROK_VERSION']
except KeyError: 
    pass
# To install the library, run the following
#
# python setup.py install
#
# prerequisite: setuptools
# http://pypi.python.org/pypi/setuptools

REQUIRES = ["urllib3 >= 1.15", "six >= 1.10", "certifi", "python-dateutil"]

setup(
    name=NAME,
    version=VERSION,
    description="zrok",
    author_email="cameron.otts@netfoundry.io",
    url="https://zrok.io",
    python_requires='>=3.10',
    keywords=["Swagger", "zrok"],
    install_requires=REQUIRES,
    packages=find_packages(),
    include_package_data=True,
    long_description="""\
    Geo-scale, next-generation peer-to-peer sharing platform built on top of OpenZiti.
    """
)
