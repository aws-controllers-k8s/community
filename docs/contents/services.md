# Services

The following AWS service APIs have service controllers included in ACK or have
controllers currently being built.

For details, including a list of planned AWS service APIs, see the [Service
Controller Release Roadmap](https://github.com/aws/aws-controllers-k8s/projects/1):

| AWS Service | Release Status | Controller |
|------------ | -------------- | ---------- |
|Amazon [API Gateway V2](https://aws.amazon.com/api-gateway/)|`DEVELOPER PREVIEW`|[`apigatewayv2`](https://github.com/aws/aws-controllers-k8s/tree/main/services/apigatewayv2)|
|Amazon [DynamoDB](https://aws.amazon.com/dynamodb/)|`DEVELOPER PREVIEW`|[`dynamodb`](https://github.com/aws/aws-controllers-k8s/tree/main/services/dynamodb)|
|Amazon [ECR](https://aws.amazon.com/ecr/)|`DEVELOPER PREVIEW`|[`ecr`](https://github.com/aws/aws-controllers-k8s/tree/main/services/ecr)|
|Amazon [S3](https://aws.amazon.com/s3/)|`DEVELOPER PREVIEW`|[`s3`](https://github.com/aws/aws-controllers-k8s/tree/main/services/s3)|
|Amazon [SQS](https://aws.amazon.com/sqs/)|`BUILD`|`sqs`|
|Amazon [SNS](https://aws.amazon.com/sns/)|`DEVELOPER PREVIEW`|[`sns`](https://github.com/aws/aws-controllers-k8s/tree/main/services/sns)|

!!! note "IMPORTANT"
    There is no single release of the ACK project. The ACK project contains a
    series of service controllers, one for each AWS service API. Each
    individual ACK service controller is released separately. Please see the
    documentation on [release criteria](releases.md) for information on how we
    release ACK service controllers.

![ACK release criteria](images/release-criteria.png)
