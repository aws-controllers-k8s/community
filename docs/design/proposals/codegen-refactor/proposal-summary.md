### Key Terms
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/model/crd.go#L63) reprensenting a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/main/models/apis/s3/2006-03-01/api-2.json#L1)
* `reconciliation`: logic involving access to or relations between `resource`, `shape`, `ackgenconfig`, and `aws-sdk`
* `ackmodel`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/model/model.go#L36) representation of ACK's view of the world; the source of truth for `aws-sdk`, `ackgenconfig`, and `reconciliation`


## Problem
*code-generator's* accelerated growth is resulting in tech debt that degrades the experience for platform maintainers and contributors. The codebase is becoming difficult to read and extend given the quick addition of features to onboard more services. As a consequence, velocity for new features and fixes have slown down and eventually contributors will become discouraged from participating altogether, if left unchecked. This doc summarizes and introduces the proposals to eliminate significant tech debt so that development on ACK Platform is as pleasant and clear as it is open.


## Solution

### Overview
To improve the implementation, I propose consolidating `ackconfig` access and any *ackconfig-aws-sdk* `reconciliation` into `ackmodel`. Taking this idea further, `ackmodel` should be the "source of truth"/required input for `api` and `controller` commands instead of *generator.yaml* and *aws-sdk*.


### Requirements
* Refactor does **not change** existing functionality
* Code is easy to read, intuitive, and extendable
* Code generation processes remain clear and transparent

## Approach

### Centralize config access and reconciliation
By centralizing and consolidating config access and `reconciliation` logic, calls becomes significantly clearer:

![current-config-access](./images/current_config_access.png)

![proposed-config-access](./images/proposed_config_access.png)


### New command `./ack-generate model`
With a new `model` command, the code generation pipeline will flow in a single direction and create opportunitities for future extension:

![proposed-gen](./images/proposed_gen.png)


## Design Proposals
The solution consists of *2 phases*; the detailed design proposals are linked below:
   * [Centralize config/model logic](./centralize_1.md)
   * [Introduce new command `model`](./model_cmd_2.md)