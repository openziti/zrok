from setuptools import setup, find_packages  # noqa: H301
import os
import versioneer
from pathlib import Path
import re

NAME = os.getenv('ZROK_PY_NAME', "zrok")
VERSION = "1.0.0"

OVERRIDES = {
    # Override specific packages with version constraints different from the generated requirements.txt
    "openziti": "openziti >= 1.0.0",
    # urllib3 2.1.0 introduced breaking changes that are implemented by openapi-generator 7.12.0
    "urllib3": "urllib3 >= 2.1.0",
}


# Parse the generated requirements.txt
def parse_requirements(filename):
    requirements = []
    if not Path(filename).exists():
        return requirements

    with open(filename, 'r') as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('#'):
                continue

            # Extract package name (everything before any version specifier)
            package_name = re.split(r'[<>=~]', line)[0].strip()

            # If we have an override for this package, skip it
            if package_name in OVERRIDES:
                continue

            requirements.append(line)

    return requirements


# Combine requirements from requirements.txt and overrides
requirements_file = Path(__file__).parent / "requirements.txt"
REQUIRES = parse_requirements(requirements_file) + list(OVERRIDES.values())

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
    package_data={
        '': ['requirements.txt'],  # Include the generated requirements.txt in the package
    },
    long_description="""\
    Geo-scale, next-generation peer-to-peer sharing platform built on top of OpenZiti.
    """
)
