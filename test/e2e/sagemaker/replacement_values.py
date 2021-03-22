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

# https://docs.aws.amazon.com/sagemaker/latest/dg/model-monitor-pre-built-container.html
MODEL_MONITOR_IMAGE_URIS = {
    "us-east-1": "156813124566.dkr.ecr.us-east-1.amazonaws.com",
    "us-east-2": "777275614652.dkr.ecr.us-east-2.amazonaws.com",
    "us-west-1": "890145073186.dkr.ecr.us-west-1.amazonaws.com",
    "us-west-2": "159807026194.dkr.ecr.us-west-2.amazonaws.com",
    "af-south-1": "875698925577.dkr.ecr.af-south-1.amazonaws.com",
    "ap-east-1": "001633400207.dkr.ecr.ap-east-1.amazonaws.com",
    "ap-northeast-1": "574779866223.dkr.ecr.ap-northeast-1.amazonaws.com",
    "ap-northeast-2": "709848358524.dkr.ecr.ap-northeast-2.amazonaws.com",
    "ap-south-1": "126357580389.dkr.ecr.ap-south-1.amazonaws.com",
    "ap-southeast-1": "245545462676.dkr.ecr.ap-southeast-1.amazonaws.com",
    "ap-southeast-2": "563025443158.dkr.ecr.ap-southeast-2.amazonaws.com",
    "ca-central-1": "536280801234.dkr.ecr.ca-central-1.amazonaws.com",
    "cn-north-1": "453000072557.dkr.ecr.cn-north-1.amazonaws.com.cn",
    "cn-northwest-1": "453252182341.dkr.ecr.cn-northwest-1.amazonaws.com.cn",
    "eu-central-1": "048819808253.dkr.ecr.eu-central-1.amazonaws.com",
    "eu-north-1": "895015795356.dkr.ecr.eu-north-1.amazonaws.com",
    "eu-south-1": "933208885752.dkr.ecr.eu-south-1.amazonaws.com",
    "eu-west-1": "468650794304.dkr.ecr.eu-west-1.amazonaws.com",
    "eu-west-2": "749857270468.dkr.ecr.eu-west-2.amazonaws.com",
    "eu-west-3": "680080141114.dkr.ecr.eu-west-3.amazonaws.com",
    "me-south-1": "607024016150.dkr.ecr.me-south-1.amazonaws.com",
    "sa-east-1": "539772159869.dkr.ecr.sa-east-1.amazonaws.com",
    "us-gov-west-1": "362178532790.dkr.ecr.us-gov-west-1.amazonaws.com",
}

# https://docs.aws.amazon.com/sagemaker/latest/dg/clarify-configure-processing-jobs.html#clarify-processing-job-configure-container
CLARIFY_IMAGE_URIS = {
    "us-east-1": "205585389593.dkr.ecr.us-east-1.amazonaws.com",
    "us-east-2": "211330385671.dkr.ecr.us-east-2.amazonaws.com",
    "us-west-1": "740489534195.dkr.ecr.us-west-1.amazonaws.com",
    "us-west-2": "306415355426.dkr.ecr.us-west-2.amazonaws.com",
    "ap-east-1": "098760798382.dkr.ecr.ap-east-1.amazonaws.com",
    "ap-south-1": "452307495513.dkr.ecr.ap-south-1.amazonaws.com",
    "ap-northeast-2": "263625296855.dkr.ecr.ap-northeast-2.amazonaws.com",
    "ap-southeast-1": "834264404009.dkr.ecr.ap-southeast-1.amazonaws.com",
    "ap-southeast-2": "007051062584.dkr.ecr.ap-southeast-2.amazonaws.com",
    "ap-northeast-1": "377024640650.dkr.ecr.ap-northeast-1.amazonaws.com",
    "ca-central-1": "675030665977.dkr.ecr.ca-central-1.amazonaws.com",
    "eu-central-1": "017069133835.dkr.ecr.eu-central-1.amazonaws.com",
    "eu-west-1": "131013547314.dkr.ecr.eu-west-1.amazonaws.com",
    "eu-west-2": "440796970383.dkr.ecr.eu-west-2.amazonaws.com",
    "eu-west-3": "341593696636.dkr.ecr.eu-west-3.amazonaws.com",
    "eu-north-1": "763603941244.dkr.ecr.eu-north-1.amazonaws.com",
    "me-south-1": "835444307964.dkr.ecr.me-south-1.amazonaws.com",
    "sa-east-1": "520018980103.dkr.ecr.sa-east-1.amazonaws.com",
    "af-south-1": "811711786498.dkr.ecr.af-south-1.amazonaws.com",
    "eu-south-1": "638885417683.dkr.ecr.eu-south-1.amazonaws.com"
}

REPLACEMENT_VALUES = {
    "SAGEMAKER_DATA_BUCKET": get_bootstrap_resources().DataBucketName,
    "XGBOOST_IMAGE_URI": f"{XGBOOST_IMAGE_URIS[get_aws_region()]}/sagemaker-xgboost:1.0-1-cpu-py3",
    "PYTORCH_TRAIN_IMAGE_URI": f"{PYTORCH_TRAIN_IMAGE_URIS[get_aws_region()]}/pytorch-training:1.5.0-cpu-py36-ubuntu16.04",
    "SAGEMAKER_EXECUTION_ROLE_ARN": get_bootstrap_resources().ExecutionRoleARN,
    "MODEL_MONITOR_ANALYZER_IMAGE_URI": f"{MODEL_MONITOR_IMAGE_URIS[get_aws_region()]}/sagemaker-model-monitor-analyzer",
    "CLARIFY_IMAGE_URI": f"{CLARIFY_IMAGE_URIS[get_aws_region()]}/sagemaker-clarify-processing:1.0"
}
