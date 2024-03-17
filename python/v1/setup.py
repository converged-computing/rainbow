import os

from setuptools import find_packages, setup  # noqa: H301

DESCRIPTION = "Python gRPC functions for the Rainbow Scheduler"

# Try to read description, otherwise fallback to short description
try:
    with open(os.path.abspath("README.md")) as filey:
        LONG_DESCRIPTION = filey.read()
except Exception:
    LONG_DESCRIPTION = DESCRIPTION

################################################################################
# MAIN #########################################################################
################################################################################

if __name__ == "__main__":
    setup(
        name="rainbow-scheduler",
        version="0.0.14rc0",
        author="Vanessasaurus",
        author_email="vsoch@users.noreply.github.com",
        maintainer="Vanessasaurus",
        packages=find_packages(),
        include_package_data=True,
        zip_safe=False,
        url="https://github.com/converged-computing/rainbow/tree/main/python/v1",
        license="MIT",
        description=DESCRIPTION,
        long_description=LONG_DESCRIPTION,
        long_description_content_type="text/markdown",
        keywords="multi-cluster, scheduler",
        setup_requires=["pytest-runner"],
        install_requires=["grpcio", "grpcio-tools", "jsonschema", "pyyaml"],
        tests_require=["pytest", "pytest-cov"],
        classifiers=[
            "Intended Audience :: Science/Research",
            "Intended Audience :: Developers",
            "License :: OSI Approved :: Apache Software License",
            "Programming Language :: C",
            "Programming Language :: Python",
            "Topic :: Software Development",
            "Topic :: Scientific/Engineering",
            "Operating System :: Unix",
            "Programming Language :: Python :: 3.7",
        ],
    )
