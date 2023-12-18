import os

from setuptools import find_packages, setup  # noqa: H301

# optionally upload to TestPyPi with alternative name in testing repo
NAME = os.getenv('ZROK_PY_NAME', "zrok_sdk")
# inherit zrok version from environment or default to dev version
VERSION = os.getenv('ZROK_VERSION', "0.4.0.dev")

# To install the library, run the following
#
# python setup.py install
#
# or
#
# pip install --editable .
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
