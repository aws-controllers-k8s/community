# `ack-generate` - Generate files for AWS Controllers for Kubernetes

## tl;dr

If you want to just build a service controller, there's a
[script](../scripts/build-controller.sh) for that:

For example, to generate the SNS service controller for ACK:

```
./scripts/build-controller.sh sns
```

## Details

The `ack-generate` tool generates files that are used in constructing
Kubernetes controllers for an individual AWS service API.

To generate a controller, we go through several generation phases:

1) Generate the type definitions for structures and enums used in custom
   resource definitions (CRDs) for the AWS service API. Part of this process is
   determining what the top-level resources are in the API. This isn't the
   easiest thing to do since all AWS service APIs are different in how they
   name API operations and objects.

   To generate type definitions, use the `ack-generate apis` command.

   ```
    ack-generate [--dry-run] apis [--version=$api_version] $service_alias
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
   ack-generate apis sns --version v1beta2 -o /tmp/ack/services/sns/v1beta2
   ```

   **NOTE**: For some APIs like the EC2 API, there will be a lot of output
   (thousands of lines). Some developers find it easier to pass the `--output`
   flag to a temporary directory and check through the generated files in that
   way instead.

2) Generate the client, deepcopy and CRD infrastructure from the type
   definitions produced in step #1:

   ```
   controller-gen object:headerFile=templates/boilerplate.txt \
     paths=./services/sns/apis/v1alpha1/...
   ```

3) Generate the controller implementation code. Every ACK service controller's
   implementation is fully generated. Use the `ack-generate controller` command to
   generate the controller implementation once you've generated the API type
   definitions in steps #1 and #2 above.

   ```
   ack-generate [--dry-run] controller $service_alias
   ```

   The `--dry-run` flag causes the command to output the controller
   implementation to `stdout`. This can be useful to check over the generated
   code before writing files to a target directory.

   When the `--dry-run` flag is false, the command writes generated files into
   multiple subdirectories under a root service controller directory (defaults
   to `services/$service_alias`). These subdirectories are `cmd/controller`
   which houses the code for the main controller binary and `pkg` which
   contains packages describing the service controller's resource managers and
   various registries.

   **NOTE**: For some APIs like the EC2 API, there will be a lot of output
   (thousands of lines). Some developers find it easier to pass the `--output`
   flag to a temporary directory and check through the generated files in that
   way instead.
