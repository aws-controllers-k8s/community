# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Handles PyTest resources and bootstrapping resource references.

PyTest resources are stored within the `resources` directory of each service
and contain YAML files used as templates for creating test fixtures.
"""

import string
import random
import yaml
import logging
from pathlib import Path
from typing import Any, Dict

from .aws import get_aws_account_id, get_aws_region

PLACEHOLDER_VALUES = {
    "AWS_ACCOUNT_ID": get_aws_account_id(),
    "AWS_REGION": get_aws_region(),
}

root_test_path = Path(__file__).parent.parent


def load_resource_file(service: str, resource_name: str,
                       additional_replacements: Dict[str, Any] = {}) -> dict:
    path = root_test_path / service / "resources"
    with open(path / f"{resource_name}.yaml", "r") as stream:
        resource_contents = stream.read()
        injected_contents = _replace_placeholder_values(
            resource_contents, PLACEHOLDER_VALUES)
        injected_contents = _replace_placeholder_values(
            injected_contents, additional_replacements)
        return yaml.safe_load(injected_contents)


def _replace_placeholder_values(
        in_str: str, replacement_dictionary: Dict[str, Any] = PLACEHOLDER_VALUES) -> str:
    for placeholder, replacement in replacement_dictionary.items():
        in_str = in_str.replace(f"${placeholder}", replacement)
    return in_str


def random_suffix_name(resource_name: str, max_length: int,
                       delimiter: str = "-") -> str:
    rand_length = max_length - len(resource_name) - len(delimiter)
    rand = "".join(random.choice(string.ascii_lowercase + string.digits)
                   for _ in range(rand_length))
    return f"{resource_name}{delimiter}{rand}"


def write_bootstrap_config(service: str, bootstrap: dict):
    path = root_test_path / service / "bootstrap.yaml"
    logging.info(f"Wrote bootstrap to {path}")
    with open(path, "w") as stream:
        yaml.safe_dump(bootstrap, stream)


def read_bootstrap_config(service: str) -> dict:
    path = root_test_path / service / "bootstrap.yaml"
    with open(path, "r") as stream:
        bootstrap = yaml.safe_load(stream)
    return bootstrap
