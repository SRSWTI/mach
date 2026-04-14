"""Setup for bodega-deploy package."""

# TODO: Future - Package enhancements:
# TODO: Migrate to pyproject.toml fully (PEP 621)
# TODO: Add optional dependencies: [docker, aws, gcp, azure]
# TODO: Implement automatic versioning from git tags (setuptools-scm)
# TODO: Add entry points for bodega plugins
# TODO: Create wheel distributions for all platforms
# TODO: Add type hints (PEP 561) and mypy checking
# TODO: Implement pre-commit hooks for code quality
# TODO: Add code coverage reporting (codecov integration)
# TODO: Sign releases with GPG for security

from setuptools import setup, find_packages
import os

# Read README
with open("README.md", "r", encoding="utf-8") as f:
    long_description = f.read()

# Package metadata
setup(
    name="bodega-deploy",
    version="0.1.0",
    author="Bodega Team",
    author_email="team@srswti.com",
    description="The simplest way to deploy anything",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/srswti/mach",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: Apache Software License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Internet :: WWW/HTTP",
        "Topic :: System :: Systems Administration",
    ],
    python_requires=">=3.8",
    install_requires=[
        "click>=8.0.0",
        "requests>=2.25.0",
        "rich>=13.0.0",
        "tomli>=2.0.0;python_version<'3.11'",
        "tomli-w>=1.0.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "black>=22.0.0",
            "flake8>=5.0.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "bodega=bodega_deploy.cli:main",
        ],
    },
    include_package_data=True,
    package_data={
        "bodega_deploy": ["bin/*"],
    },
    zip_safe=False,
)
