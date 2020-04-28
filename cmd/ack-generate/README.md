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
   accepts (via `stdin` or as the first positional argument) a JSON or YAML
   file containing an OpenAPI3 Schema document for the AWS service API you wish
   to generate type definitions and basic scaffolding for:

   ```
    ack-generate types < /path/to/schema.yaml
   ```

   The command outputs the type definitions, enumerations and basic type
   registration scaffolding to `stdout`. This is useful to check over the
   generated files before writing files to a target directory.

   To write individual files, use the `-o|--output` flag, which accepts a path
   to the directory you want to send generated files to. If this directory does
   not exist, the `ack-generate types` command will ensure it exists:

   ```
   ack-generate types -o services/sns/v1alpha1 < /tmp/sns.yaml
   ```
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
