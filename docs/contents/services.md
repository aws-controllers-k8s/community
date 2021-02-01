# Services

The following AWS service APIs have service controllers included in ACK or have
controllers currently being built.

For details, including a list of planned AWS service APIs, see the [Service
Controller Release Roadmap](https://github.com/aws/aws-controllers-k8s/projects/1):

!!! note "IMPORTANT"
    There is no single release of the ACK project. The ACK project contains a
    series of service controllers, one for each AWS service API. Each
    individual ACK service controller is released separately. Please see the
    documentation on [release criteria](releases.md) for information on how we
    release ACK service controllers.

| AWS Service | Current Status | Next Milestone
| ----------- | -------------- | --------------
| Amazon [ACM](#amazon-acm) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/482) |
| Amazon [API Gateway V2](#amazon-api-gateway-v2) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/207) | [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/15)
| Amazon [CloudFront Distribution](#amazon-cloudfront-distribution) | [`PLANNED`](https://github.com/aws/aws-controllers-k8s/issues/249) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/14)
| Amazon [DynamoDB](#amazon-dynamodb) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/2060) |
| Amazon [ECR](#amazon-ecr) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/208) |
| Amazon [EFS](#amazon-efs) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/328) |
| Amazon [EKS](#amazon-eks) | [`PLANNED`](https://github.com/aws/aws-controllers-k8s/issues/16) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/7)
| Amazon [ElastiCache](#amazon-elasticache) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/240) | [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/9)
| Amazon [Elasticsearch](#amazon-elasticsearch) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/503) |
| Amazon [EC2 VPC](#amazon-ec2-vpc) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/489) |
| AWS [IAM](#aws-iam) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/222) |
| AWS [Lambda](#aws-lambda) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/238) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/10)
| AWS [Kinesis](#aws-kinesis) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/235) |
| Amazon [KMS](#amazon-kms) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/491) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/18)
| Amazon [MQ](#amazon-mq) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/390) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/12)
| Amazon [MSK](#amazon-msk) | [`PLANNED`](https://github.com/aws/aws-controllers-k8s/issues/348) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/13)
| Amazon [RDS](#amazon-rds) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/237) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/8)
| Amazon [Route53](#amazon-route53) | [`PROPOSED`](https://github.com/aws/aws-controllers-k8s/issues/480) |
| Amazon [SageMaker](#amazon-sagemaker) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/385) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/11)
| Amazon [SNS](#amazon-sns) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/202) |
| Amazon [SQS](#amazon-sqs) | [`BUILD`](https://github.com/aws/aws-controllers-k8s/issues/205) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/6)
| AWS [Step Functions](#aws-step-functions) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/239) |
| Amazon [S3](#amazon-s3) | [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/issues/204) |

!!! note "Don't see a service listed?"
    If you don't see a particular AWS service listed, feel free to
    [propose it](https://github.com/aws/aws-controllers-k8s/issues/new?labels=Service+Controller&template=propose_new_controller.md&title=%5Bname%5D+service+controller)!

## Amazon ACM

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/482
* Current release status: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/acm/

## Amazon API Gateway v2

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/207
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/15)
* AWS service documentation: https://aws.amazon.com/api-gateway/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/apigatewayv2

## Amazon CloudFront Distribution

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/249
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/14)
* Current release status: `PLANNED`
* AWS service documentation: https://aws.amazon.com/cloudfront/

## Amazon DynamoDB

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/206
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/16)
* AWS service documentation: https://aws.amazon.com/dynamodb/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/dynamodb

## Amazon ECR

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/208
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/16)
* AWS service documentation: https://aws.amazon.com/ecr/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/ecr

## Amazon EFS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/328
* Current release status: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/efs/

## Amazon EKS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/16
* Current release status: `PLANNED`
* AWS service documentation: https://aws.amazon.com/eks/

## Amazon ElastiCache

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/240
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/9)
* AWS service documentation: https://aws.amazon.com/elasticache/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/elasticache

## Amazon Elasticsearch

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/503
* Current release status: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/elasticsearch-service/

## Amazon EC2 VPC

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/489
* Current release status: `PROPOSED`
* AWS service documentation: https://docs.aws.amazon.com/vpc/

## AWS IAM

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/222
* Current release status: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/iam/

## AWS Lambda

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/238
* Current release status: `BUILD`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/10)
* AWS service documentation: https://aws.amazon.com/lambda/

## Amazon Kinesis

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/235
* Current release status: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/kinesis/

## AWS KMS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/491
* Current release status: `BUILD`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/18)
* AWS service documentation: https://aws.amazon.com/kms/

## Amazon MQ

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/390
* Current release status: `BUILD`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/12)
* AWS service documentation: https://aws.amazon.com/amazon-mq/

## Amazon MSK

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/348
* Current release status: `PLANNED`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/13)
* AWS service documentation: https://aws.amazon.com/msk/

## Amazon RDS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/237
* Current release status: `PLANNED`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/8)
* AWS service documentation: https://aws.amazon.com/rds/

## Amazon Route53

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/480
* Current release status: `PROPOSED`
* AWS service documentation: https://docs.aws.amazon.com/Route53/

## Amazon SageMaker

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/385
* Current release status: `BUILD`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/11)
* AWS service documentation: https://aws.amazon.com/sagemaker/

## Amazon SNS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/202
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/17)
* AWS service documentation: https://aws.amazon.com/sns/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/sns

## Amazon SQS

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/205
* Current release status: `BUILD`
* Next milestone: [`DEVELOPER PREVIEW`](https://github.com/aws/aws-controllers-k8s/milestone/6)
* AWS service documentation: https://aws.amazon.com/sqs/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/sqs

## AWS Step Functions

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/239
* Current release status: `DEVELOPER PREVIEW`
* AWS service documentation: https://aws.amazon.com/step-functions/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/sfn

## Amazon S3

* Proposed: https://github.com/aws/aws-controllers-k8s/issues/204
* Current release status: `DEVELOPER PREVIEW`
* Next milestone: [`BETA`](https://github.com/aws/aws-controllers-k8s/milestone/16)
* AWS service documentation: https://aws.amazon.com/s3/
* ACK service controller: https://github.com/aws/aws-controllers-k8s/tree/main/services/s3
