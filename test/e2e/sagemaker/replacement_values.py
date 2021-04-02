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
"""Stores the values used by each of the integration tests for replacing the
SageMaker-specific test variables.
"""

from common.aws import get_aws_region
from sagemaker.bootstrap_resources import get_bootstrap_resources

# Taken from the SageMaker Python SDK
# Rather than including the entire SDK
XGBOOST_IMAGE_URIS = {
    "us-west-1": 	    "746614075791.dkr.ecr.us-west-1.amazonaws.com",
    "us-west-2": 	    "246618743249.dkr.ecr.us-west-2.amazonaws.com",
    "us-east-1": 	    "683313688378.dkr.ecr.us-east-1.amazonaws.com",
    "us-east-2": 	    "257758044811.dkr.ecr.us-east-2.amazonaws.com",
    "ap-east-1": 	    "651117190479.dkr.ecr.ap-east-1.amazonaws.com",
    "ap-northeast-1": 	"354813040037.dkr.ecr.ap-northeast-1.amazonaws.com",
    "ap-northeast-2": 	"366743142698.dkr.ecr.ap-northeast-2.amazonaws.com",
    "ap-south-1": 	    "720646828776.dkr.ecr.ap-south-1.amazonaws.com",
    "ap-southeast-1": 	"121021644041.dkr.ecr.ap-southeast-1.amazonaws.com",
    "ap-southeast-2": 	"783357654285.dkr.ecr.ap-southeast-2.amazonaws.com",
    "ca-central-1": 	"341280168497.dkr.ecr.ca-central-1.amazonaws.com",
    "cn-north-1": 	    "450853457545.dkr.ecr.cn-north-1.amazonaws.com.cn",
    "cn-northwest-1": 	"451049120500.dkr.ecr.cn-northwest-1.amazonaws.com.cn",
    "eu-central-1": 	"492215442770.dkr.ecr.eu-central-1.amazonaws.com",
    "eu-north-1": 	    "662702820516.dkr.ecr.eu-north-1.amazonaws.com",
    "eu-west-1": 	    "141502667606.dkr.ecr.eu-west-1.amazonaws.com",
    "eu-west-2": 	    "764974769150.dkr.ecr.eu-west-2.amazonaws.com",
    "eu-west-3": 	    "659782779980.dkr.ecr.eu-west-3.amazonaws.com",
    "me-south-1": 	    "801668240914.dkr.ecr.me-south-1.amazonaws.com",
    "sa-east-1": 	    "737474898029.dkr.ecr.sa-east-1.amazonaws.com"
}

XGBOOST_DEBUGGER_IMAGE_URIS = {
    "us-west-1": 	    "685455198987.dkr.ecr.us-west-1.amazonaws.com",
    "us-west-2": 	    "895741380848.dkr.ecr.us-west-2.amazonaws.com",
    "us-east-1": 	    "503895931360.dkr.ecr.us-east-1.amazonaws.com",
    "us-east-2": 	    "915447279597.dkr.ecr.us-east-2.amazonaws.com",
    "ap-east-1": 	    "199566480951.dkr.ecr.ap-east-1.amazonaws.com",
    "ap-northeast-1": 	"430734990657.dkr.ecr.ap-northeast-1.amazonaws.com",
    "ap-northeast-2": 	"578805364391.dkr.ecr.ap-northeast-2.amazonaws.com",
    "ap-south-1": 	    "904829902805.dkr.ecr.ap-south-1.amazonaws.com",
    "ap-southeast-1": 	"972752614525.dkr.ecr.ap-southeast-1.amazonaws.com",
    "ap-southeast-2": 	"184798709955.dkr.ecr.ap-southeast-2.amazonaws.com",
    "ca-central-1": 	"519511493484.dkr.ecr.ca-central-1.amazonaws.com",
    "cn-north-1": 	    "618459771430.dkr.ecr.cn-north-1.amazonaws.com.cn",
    "cn-northwest-1": 	"658757709296.dkr.ecr.cn-northwest-1.amazonaws.com.cn",
    "eu-central-1": 	"482524230118.dkr.ecr.eu-central-1.amazonaws.com",
    "eu-north-1": 	    "314864569078.dkr.ecr.eu-north-1.amazonaws.com",
    "eu-west-1": 	    "929884845733.dkr.ecr.eu-west-1.amazonaws.com",
    "eu-west-2": 	    "250201462417.dkr.ecr.eu-west-2.amazonaws.com",
    "eu-west-3": 	    "447278800020.dkr.ecr.eu-west-3.amazonaws.com",
    "me-south-1": 	    "986000313247.dkr.ecr.me-south-1.amazonaws.com",
    "sa-east-1": 	    "818342061345.dkr.ecr.sa-east-1.amazonaws.com"
}

PYTORCH_TRAIN_IMAGE_URIS = {
    "us-east-1":        "763104351884.dkr.ecr.us-east-1.amazonaws.com",
    "us-east-2":        "763104351884.dkr.ecr.us-east-2.amazonaws.com",
    "us-west-1":        "763104351884.dkr.ecr.us-west-1.amazonaws.com",
    "us-west-2":        "763104351884.dkr.ecr.us-west-2.amazonaws.com",
    "af-south-1":       "626614931356.dkr.ecr.af-south-1.amazonaws.com",
    "ap-east-1":        "871362719292.dkr.ecr.ap-east-1.amazonaws.com",
    "ap-south-1":       "763104351884.dkr.ecr.ap-south-1.amazonaws.com",
    "ap-northeast-2":   "763104351884.dkr.ecr.ap-northeast-2.amazonaws.com",
    "ap-southeast-1":   "763104351884.dkr.ecr.ap-southeast-1.amazonaws.com",
    "ap-southeast-2":   "763104351884.dkr.ecr.ap-southeast-2.amazonaws.com",
    "ap-northeast-1":   "763104351884.dkr.ecr.ap-northeast-1.amazonaws.com",
    "ca-central-1":     "763104351884.dkr.ecr.ca-central-1.amazonaws.com",
    "eu-central-1":     "763104351884.dkr.ecr.eu-central-1.amazonaws.com",
    "eu-west-1":        "763104351884.dkr.ecr.eu-west-1.amazonaws.com",
    "eu-west-2":        "763104351884.dkr.ecr.eu-west-2.amazonaws.com",
    "eu-south-1":       "692866216735.dkr.ecr.eu-south-1.amazonaws.com",
    "eu-west-3":        "763104351884.dkr.ecr.eu-west-3.amazonaws.com",
    "eu-north-1":       "763104351884.dkr.ecr.eu-north-1.amazonaws.com",
    "me-south-1":       "217643126080.dkr.ecr.me-south-1.amazonaws.com",
    "sa-east-1":        "763104351884.dkr.ecr.sa-east-1.amazonaws.com",
    "cn-north-1":       "727897471807.dkr.ecr.cn-north-1.amazonaws.com.cn",
    "cn-northwest-1":   "727897471807.dkr.ecr.cn-northwest-1.amazonaws.com.cn"
}

REPLACEMENT_VALUES = {
    "SAGEMAKER_DATA_BUCKET": get_bootstrap_resources().DataBucketName,
    "XGBOOST_IMAGE_URI": f"{XGBOOST_IMAGE_URIS[get_aws_region()]}/sagemaker-xgboost:1.0-1-cpu-py3",
    "XGBOOST_DEBUGGER_IMAGE_URI": f"{XGBOOST_DEBUGGER_IMAGE_URIS[get_aws_region()]}/sagemaker-debugger-rules:latest",
    "PYTORCH_TRAIN_IMAGE_URI": f"{PYTORCH_TRAIN_IMAGE_URIS[get_aws_region()]}/pytorch-training:1.5.0-cpu-py36-ubuntu16.04",
    "SAGEMAKER_EXECUTION_ROLE_ARN": get_bootstrap_resources().ExecutionRoleARN
}
