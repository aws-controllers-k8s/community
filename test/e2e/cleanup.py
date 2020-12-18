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
"""Starts the clean up process for the bootstrapped service resources for the
selected service.
"""

import sys
from pathlib import Path
from importlib import import_module

from common.resources import read_bootstrap_config

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(f"{__file__} requires a single parameter: <service>")
        sys.exit(1)

    service_name = sys.argv[1]
    import importlib.util

    # TODO(nithomso): Investigate how to move this to importlib
    # I've spent 3+ hours trying, but I'm sure there's a way
    service_cleanup = __import__(f"{service_name}.service_cleanup").service_cleanup

    bootstrap_config = read_bootstrap_config(service_name)
    service_cleanup.service_cleanup(bootstrap_config)