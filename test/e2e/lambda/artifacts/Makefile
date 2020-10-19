LAMBDA_FUNCTION_NAME ?= ack-test
IAM_ID ?= 000000000000
S3_BUCKET ?= ack-testing-lambda
AWS_REGION ?= eu-west-2

build:
	GOOS=linux GOARCH=amd64 go build -o lambda lambda.go
	zip -j lambda.zip lambda

create-s3bucket:
	aws s3api create-bucket \
		--bucket=$(S3_BUCKET) \
		--region $(region) \
		--create-bucket-configuration LocationConstraint=$(region)

upload:
	aws s3 cp ./lambda.zip s3://$(S3_BUCKET)/lambda.zip

clean:
	aws lambda delete-function \
		--function-name $(LAMBDA_FUNCTION_NAME)

all: build create-s3bucket upload

.PHONY: build upload all clean