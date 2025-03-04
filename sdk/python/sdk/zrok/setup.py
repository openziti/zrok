from setuptools import setup, find_packages  # noqa: H301
import os
import versioneer

# optionally upload to TestPyPi with alternative name in testing repo
NAME = os.getenv('ZROK_PY_NAME', "zrok")
VERSION = "1.0.0"
REQUIRES = [
    "openziti >= 1.0.0",
    "urllib3 >= 2.1.0",  # urllib3 2.1.0 introduced breaking changes that are implemented by openapi-generator 7.12.0
    "python_dateutil >= 2.8.2",
    "pydantic >= 2",
    "typing-extensions >= 4.7.1",
]

setup(
    name=NAME,
    cmdclass=versioneer.get_cmdclass(dict()),
    version=versioneer.get_version(),
    description="zrok",
    author_email="",
    url="",
    keywords=["zrok"],
    install_requires=REQUIRES,
    python_requires='>3.10.0',
    packages=find_packages(),
    include_package_data=True,
    long_description="""\
    Geo-scale, next-generation peer-to-peer sharing platform built on top of OpenZiti.
    """
)
