LAMBDA_FUNCTION_NAME ?= ack-test
IAM_ID ?= 771174509839
S3_BUCKET ?= ack-testing-lambda-98756

build:
	GOOS=linux GOARCH=amd64 go build -o lambda lambda.go
	zip -j lambda.zip lambda

create-s3bucket:
	aws s3api create-bucket \
		--bucket=$(S3_BUCKET) \
		--region eu-west-2 \
		--create-bucket-configuration LocationConstraint=eu-west-2

upload:
	aws s3 cp ./lambda.zip s3://$(S3_BUCKET)/lambda.zip

# Doesn't use ack
deploy:
	aws lambda create-function \
		--runtime go1.x \
		--handler lambda \
		--function-name $(LAMBDA_FUNCTION_NAME) \
		--role arn:aws:iam::$(IAM_ID):role/lambda-executor \
		--zip-file fileb://$(PWD)/lambda.zip

# Doesn't use ack
update:
    aws lambda update-function-code \
        --function-name $(LAMBDA_FUNCTION_NAME) \
        --zip-file fileb://$(PWD)/lambda.zip

invoke:
    aws lambda invoke \
        --function-name $(LAMBDA_FUNCTION_NAME) /tmp/output.json \
        --payload '{"name":"test-event"}'
	cat /tmp/output.json | jq

clean:
	aws lambda delete-function \
		--function-name $(LAMBDA_FUNCTION_NAME)
 
all: build update invoke
new: build deploy invoke

.PHONY: build deploy update invoke all new