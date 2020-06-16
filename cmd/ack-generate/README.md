# `ack-generate` - Generate files for AWS Controllers for Kubernetes

The `ack-generate` tool generates files that are used in constructing
Kubernetes controllers for an individual AWS service API.

To generate a controller, we go through several generation phases:

1) Generate the type definitions for structures and enums used in custom
   resource definitions (CRDs) for the AWS service API. Part of this process is
   determining what the top-level resources are in the API. This isn't the
   easiest thing to do since all AWS service APIs are different in how they
   name API operations and objects.

   To generate type definitions, use the `ack-generate types` command. It
   accepts (via `stdin` or the `-i|--input` flag) a JSON or YAML file
   containing an OpenAPI3 Schema document for the AWS service API you wish to
   generate type definitions and basic scaffolding for:

   ```
    ack-generate [--dry-run] types [--version=$api_version] $service_alias < /path/to/schema.yaml
   ```

   The `--dry-run` flag causes the command to output the type definitions,
   enumerations and basic type registration scaffolding to `stdout`. This is
   useful to check over the generated files before writing files to a target
   directory.

   When the `--dry-run` flag is false, the command writes generated files to a
   directory (defaults to `services/$service_alias/apis/$api_version`). To
   override the directory that files are written to,  use the `-o|--output`
   flag, which accepts a path to the directory you want to send generated files
   to. If this directory does not exist, the `ack-generate types` command will
   ensure it exists:

   ```
   ack-generate types sns --version v1beta2 -o /tmp/ack/services/sns/v1beta2 < /tmp/sns.yaml
   ```

   **NOTE**: For some APIs like the EC2 API, there will be a lot of output
   (thousands of lines). Some developers find it easier to pass the `--output`
   flag to a temporary directory and check through the generated files in that
   way instead.

2) Generate the client, deepcopy and CRD infrastructure from the type
   definitions produced in step #1 (TODO)

3) Generate the controller implementation code (TODO)

## Get an OpenAPI3 Schema document for an AWS service API

Don't have an OpenAPI3 Schema document for a particular AWS service API? Not to
worry. An easy way to get one is to use the
[`aws-api-tool schema <api>` command](https://github.com/jaypipes/aws-api-tools#show-openapi3-schema-swagger-for-api):

For example, to generate the OpenAPI3 Schema document for the Amazon Elastic
Kubernetes Service (EKS) API, do:

```
aws-api-tool schema eks > /tmp/eks.yaml
```

Read more about how to install and use `aws-api-tool` [here](https://github.com/jaypipes/aws-api-tools).
