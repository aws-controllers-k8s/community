# End-to-End Testing
The ACK End-to-End (E2E) testing framework intends to test each of the service
operators against a set of known custom resource definitions.

## Overview
The overall flow for the automated testing any one service is as follows:
1. Run a bootstrap Python script
    - Invokes the individual service's `service_bootstrap.py` file,
        creating AWS resources necessary to support the custom resource tests
    - Stores the exported properties of each resource into a `bootstrap.yaml`
        file in the service directory

2. Run the PyTest module in the service directory
    - Helper methods load templated resources (custom resource YAMLs) and
        injects values
    - Runs the PyTest fixtures and tests as expected

3. Run a cleanup Python script
    - Invokes the individual services `service_cleanup.py` file, deleting the
        bootstrapped resources

Each of the steps can also be
[run independently](#manual-invocation-for-local-development), for the sake of
faster debugging. Otherwise, a Dockerfile is provided which encapsulates the
testing procedure and includes all necessary requirements.

## Fully Automated Invocation
The E2E tests can be run entirely within Docker. The tests require that you have
a configured `~/.kubeconfig/config` file and exported AWS environment variables.
The container will run all of the tests as outlined in the
[overview](#overview).

### Requirements
- [Docker 18.06+](https://www.docker.com/)

### Running the Tests
To invoke the automated tests, navigate to the `test/e2e` directory in your 
shell and then invoke the `build-run-test-dockerfile.sh` file followed by the 
name of the service you wish to test.
For example:
```bash
./build-run-test-dockerfile.sh s3
```


## Manual Invocation (for Local Development)
Manual invocation allows developers to run any part of the automated test flow
on their host machine, rather than in a Docker container. This could be 
preferred over fully automated testing as that process bootstraps and cleans up
AWS resources on every invocation. For some services, this could mean multiple
minutes of waiting for resources to reach their ready state before actually 
running any tests.

### Requirements
For local testing and development, the requirements are:
- Python 3.8+
- Conda (or PyEnv)

Activate your Python environment and install the `pip` packages as such:
```bash
pip install -r requirements.txt
```

### Running the Tests
To bootstrap a service's resources:
```bash
python ./bootstrap.py <service_name>
```

At the conclusion of bootstrapping, you should see the YAML containing the 
bootstrapped resource definitions (as defined by the services bootstrap class)
in the service directory with the file name `bootstrap.yaml`.

To run a service's tests:
```bash
cd <service_name>
pytest --log-cli-level INFO .
```

To clean up a service's bootstrapped resources:
```bash
python ./cleanup.py <service_name>
```