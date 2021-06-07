# Services

The following AWS service APIs have service controllers included in ACK or have
controllers in one of our [several project stages][project-stages].

[project-stages]: https://aws-controllers-k8s.github.io/community/releases#project-stages

ACK controllers that have reached the `RELEASED` project stage will also be in
one of our [maintenance phases][maint-phases].

[maint-phases]: https://aws-controllers-k8s.github.io/community/releases#maintenance-phases

For details, including a list of planned AWS service APIs, see the [Service
Controller Release Roadmap](https://github.com/aws-controllers-k8s/community/projects/1):

!!! note "IMPORTANT"
    There is no single release of the ACK project. The ACK project contains a
    series of service controllers, one for each AWS service API. Each
    individual ACK service controller is released separately. Please see the
    [release documentation][releases] for information on how we version and
    release ACK service controllers.

[releases]: https://aws-controllers-k8s.github.io/community/releases

| AWS Service | Project Stage | Maintenance Phase | Next Milestone 
| ----------- | ------------- | ----------------- | -------------- 
| Amazon [ACM](#amazon-acm) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/482) | |
| Amazon [API Gateway V2](#amazon-api-gateway-v2) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/207) | `PREVIEW` | https://github.com/aws-controllers-k8s/community/milestone/15
| Amazon [Application Auto Scaling](#amazon-application-auto-scaling) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/589) | `PREVIEW` |
| Amazon [CloudFront Distribution](#amazon-cloudfront-distribution) | [`PLANNED`](https://github.com/aws-controllers-k8s/community/issues/249) | | https://github.com/aws-controllers-k8s/community/milestone/14
| Amazon [DynamoDB](#amazon-dynamodb) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/2060) | `PREVIEW` |
| Amazon [ECR](#amazon-ecr) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/208) | `PREVIEW` |
| Amazon [EFS](#amazon-efs) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/328) | |
| Amazon [EKS](#amazon-eks) | [`PLANNED`](https://github.com/aws-controllers-k8s/community/issues/16) | | https://github.com/aws-controllers-k8s/community/milestone/7
| Amazon [ElastiCache](#amazon-elasticache) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/240) | `PREVIEW` | https://github.com/aws-controllers-k8s/community/milestone/9
| Amazon [Elasticsearch](#amazon-elasticsearch) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/503) | |
| Amazon [EC2 VPC](#amazon-ec2-vpc) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/489) | |
| AWS [IAM](#aws-iam) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/222) | |
| AWS [Lambda](#aws-lambda) | [`IN PROGRESS`](https://github.com/aws-controllers-k8s/community/issues/238) | | https://github.com/aws-controllers-k8s/community/milestone/10
| AWS [Kinesis](#aws-kinesis) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/235) | |
| Amazon [KMS](#amazon-kms) | [`IN PROGRESS`](https://github.com/aws-controllers-k8s/community/issues/491) | | https://github.com/aws-controllers-k8s/community/milestone/18
| Amazon [MQ](#amazon-mq) | [`IN PROGRESS`](https://github.com/aws-controllers-k8s/community/issues/390) | | https://github.com/aws-controllers-k8s/community/milestone/12
| Amazon [MSK](#amazon-msk) | [`PLANNED`](https://github.com/aws-controllers-k8s/community/issues/348) | | https://github.com/aws-controllers-k8s/community/milestone/13
| Amazon [RDS](#amazon-rds) | [`IN PROGRESS`](https://github.com/aws-controllers-k8s/community/issues/237) | | https://github.com/aws-controllers-k8s/community/milestone/8
| Amazon [Route53](#amazon-route53) | [`PROPOSED`](https://github.com/aws-controllers-k8s/community/issues/480) | |
| Amazon [SageMaker](#amazon-sagemaker) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/385) | `PREVIEW` | https://github.com/aws-controllers-k8s/community/milestone/11
| Amazon [SNS](#amazon-sns) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/202) | `PREVIEW` |
| Amazon [SQS](#amazon-sqs) | [`IN PROGRESS`](https://github.com/aws-controllers-k8s/community/issues/205) | | https://github.com/aws-controllers-k8s/community/milestone/6
| AWS [Step Functions](#aws-step-functions) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/239) | `PREVIEW` |
| Amazon [S3](#amazon-s3) | [`RELEASED`](https://github.com/aws-controllers-k8s/community/issues/204) | `PREVIEW` |

!!! note "Don't see a service listed?"
    If you don't see a particular AWS service listed, feel free to
    [propose it](https://github.com/aws-controllers-k8s/community/issues/new?labels=Service+Controller&template=propose_new_controller.md&title=%5Bname%5D+service+controller)!

## Amazon ACM

* Proposed: https://github.com/aws-controllers-k8s/community/issues/482
* Current project stage: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/acm/

## Amazon API Gateway v2

* Proposed: https://github.com/aws-controllers-k8s/community/issues/207
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/15
* AWS service documentation: https://aws.amazon.com/api-gateway/
* ACK service controller: https://github.com/aws-controllers-k8s/apigatewayv2-controller

## Amazon Application Auto Scaling

* Proposed: https://github.com/aws-controllers-k8s/community/issues/589
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* AWS service documentation: https://docs.aws.amazon.com/autoscaling/application/userguide/what-is-application-auto-scaling.html
* ACK service controller: https://github.com/aws-controllers-k8s/applicationautoscaling-controller

## Amazon CloudFront Distribution

* Proposed: https://github.com/aws-controllers-k8s/community/issues/249
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/14
* Current project stage: `PLANNED`
* AWS service documentation: https://aws.amazon.com/cloudfront/

## Amazon DynamoDB

* Proposed: https://github.com/aws-controllers-k8s/community/issues/206
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/16
* AWS service documentation: https://aws.amazon.com/dynamodb/
* ACK service controller: https://github.com/aws-controllers-k8s/dynamodb-controller

## Amazon ECR

* Proposed: https://github.com/aws-controllers-k8s/community/issues/208
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/16
* AWS service documentation: https://aws.amazon.com/ecr/
* ACK service controller: https://github.com/aws-controllers-k8s/ecr-controller

## Amazon EFS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/328
* Current project stage: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/efs/

## Amazon EKS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/16
* Current project stage: `PLANNED`
* AWS service documentation: https://aws.amazon.com/eks/

## Amazon ElastiCache

* Proposed: https://github.com/aws-controllers-k8s/community/issues/240
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/9
* AWS service documentation: https://aws.amazon.com/elasticache/
* ACK service controller: https://github.com/aws-controllers-k8s/elasticache-controller

## Amazon Elasticsearch

* Proposed: https://github.com/aws-controllers-k8s/community/issues/503
* Current project stage: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/elasticsearch-service/

## Amazon EC2 VPC

* Proposed: https://github.com/aws-controllers-k8s/community/issues/489
* Current project stage: `PROPOSED`
* AWS service documentation: https://docs.aws.amazon.com/vpc/

## AWS IAM

* Proposed: https://github.com/aws-controllers-k8s/community/issues/222
* Current project stage: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/iam/

## AWS Lambda

* Proposed: https://github.com/aws-controllers-k8s/community/issues/238
* Current project stage: `IN PROGRESS`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/10
* AWS service documentation: https://aws.amazon.com/lambda/

## Amazon Kinesis

* Proposed: https://github.com/aws-controllers-k8s/community/issues/235
* Current project stage: `PROPOSED`
* AWS service documentation: https://aws.amazon.com/kinesis/

## AWS KMS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/491
* Current project stage: `IN PROGRESS`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/18
* AWS service documentation: https://aws.amazon.com/kms/
* ACK service controller: https://github.com/aws-controllers-k8s/kms-controller

## Amazon MQ

* Proposed: https://github.com/aws-controllers-k8s/community/issues/390
* Current project stage: `IN PROGRESS`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/12
* AWS service documentation: https://aws.amazon.com/amazon-mq/
* ACK service controller: https://github.com/aws-controllers-k8s/mq-controller

## Amazon MSK

* Proposed: https://github.com/aws-controllers-k8s/community/issues/348
* Current project stage: `PLANNED`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/13
* AWS service documentation: https://aws.amazon.com/msk/

## Amazon RDS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/237
* Current project stage: `PLANNED`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/8
* AWS service documentation: https://aws.amazon.com/rds/

## Amazon Route53

* Proposed: https://github.com/aws-controllers-k8s/community/issues/480
* Current project stage: `PROPOSED`
* AWS service documentation: https://docs.aws.amazon.com/Route53/

## Amazon SageMaker

* Proposed: https://github.com/aws-controllers-k8s/community/issues/385
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/11
* AWS service documentation: https://aws.amazon.com/sagemaker/
* ACK service controller: https://github.com/aws-controllers-k8s/sagemaker-controller

## Amazon SNS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/202
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/17
* AWS service documentation: https://aws.amazon.com/sns/
* ACK service controller: https://github.com/aws-controllers-k8s/sns-controller

## Amazon SQS

* Proposed: https://github.com/aws-controllers-k8s/community/issues/205
* Current project stage: `IN PROGRESS`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/6
* AWS service documentation: https://aws.amazon.com/sqs/
* ACK service controller: https://github.com/aws-controllers-k8s/sqs-controller

## AWS Step Functions

* Proposed: https://github.com/aws-controllers-k8s/community/issues/239
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* AWS service documentation: https://aws.amazon.com/step-functions/
* ACK service controller: https://github.com/aws-controllers-k8s/sfn-controller

## Amazon S3

* Proposed: https://github.com/aws-controllers-k8s/community/issues/204
* Current project stage: `RELEASED`
* Current maintenance phase: `PREVIEW`
* Next milestone: https://github.com/aws-controllers-k8s/community/milestone/16
* AWS service documentation: https://aws.amazon.com/s3/
* ACK service controller: https://github.com/aws-controllers-k8s/s3-controller
