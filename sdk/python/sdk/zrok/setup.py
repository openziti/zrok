import os

from setuptools import find_packages, setup  # noqa: H301

import versioneer

# optionally upload to TestPyPi with alternative name in testing repo
NAME = os.getenv('ZROK_PY_NAME', "zrok")

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

REQUIRES = ["urllib3 >= 1.15", "six >= 1.10", "certifi", "python-dateutil", "openziti >= 0.8.1"]

setup(
    name=NAME,
    cmdclass=versioneer.get_cmdclass(dict()),
    version=versioneer.get_version(),
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
