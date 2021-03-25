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
"""ElastiCache-specific test utility functions
"""

import logging
import boto3

from time import sleep

ec = boto3.client("elasticache")

def wait_usergroup_active(usergroup_id: str,
                           wait_periods: int = 10,
                           period_length: int = 60) -> bool:
    for i in range(wait_periods):
        logging.debug(f"Waiting for user group {usergroup_id} to be active ({i})")
        response = ec.describe_user_groups(UserGroupId=usergroup_id)
        user_group = response['UserGroups'][0]

        if not user_group:
            logging.error(f"Could not find User Group {usergroup_id}")
            return False

        if user_group['Status'] == "active":
            logging.info(f"User Group {usergroup_id} is active, continuing...")
            return True

        sleep(period_length)

    logging.error(f"Wait for User Group {usergroup_id} to be active timed out")
    return False

def wait_snapshot_available(snapshot_name: str,
                            wait_periods: int = 10,
                            period_length: int = 60) -> bool:
    for i in range(wait_periods):
        logging.debug(f"Waiting for snapshot {snapshot_name} to be available ({i})")
        response = ec.describe_snapshots(SnapshotName=snapshot_name)
        snapshot = response['Snapshots'][0]

        if not snapshot:
            logging.error(f"Could not find snapshot {snapshot_name}")
            return False

        if snapshot['SnapshotStatus'] == "available":
            logging.info(f"Snapshot {snapshot_name} is available, continuing...")
            return True

        sleep(period_length)

    logging.error(f"Wait for snapshot {snapshot_name} to be available timed out")
    return False

def wait_snapshot_deleted(snapshot_name: str,
                          wait_periods: int = 10,
                          period_length: int = 60) -> bool:
    for i in range(wait_periods):
        logging.debug(f"Waiting for snapshot {snapshot_name} to be deleted ({i})")
        response = ec.describe_snapshots(SnapshotName=snapshot_name)

        if len(response['Snapshots']) == 0:
            return True

        sleep(period_length)

    logging.error(f"Wait for snapshot {snapshot_name} to be deleted timed out")
    return False

# provide a basic nodeGroupConfiguration object of desired size
def provide_node_group_configuration(size: int):
    ngc = []
    for i in range(1, size+1):
        ngc.append({"nodeGroupID": str(i).rjust(4, '0')})
    return ngc