from setuptools import setup, find_packages  # noqa: H301
import os
import versioneer

NAME = os.getenv('ZROK_PY_NAME') or "zrok2"
VERSION = "2.0.0"

REQUIRES = [
    "openziti >= 1.0.0",
    "urllib3 >= 2.1.0",
    "pydantic >= 2",
    "python_dateutil >= 2.8.2",
    "typing-extensions >= 4.7.1",
]

setup(
    name=NAME,
    cmdclass=versioneer.get_cmdclass(dict()),
    version=versioneer.get_version(),
    description="zrok2",
    author_email="",
    url="",
    keywords=["zrok", "zrok2"],
    install_requires=REQUIRES,
    extras_require={
        "test": ["pytest>=7.0", "pytest-cov"],
    },
    python_requires='>3.10.0',
    packages=find_packages(),
    include_package_data=True,
    package_data={
        '': ['requirements.txt'],  # Include the generated requirements.txt in the package
    },
    long_description="""\
    Geo-scale, next-generation peer-to-peer sharing platform built on top of OpenZiti.
    """
)
