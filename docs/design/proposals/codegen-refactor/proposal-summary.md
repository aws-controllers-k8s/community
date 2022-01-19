### Key Terms
* `ACK platforn` | `platform`: the set of code generators, test frameworks, and model inference utilities that comprise the AWS Controllers for Kubernetes project.
* `pipeline`: the collection of all phases involved in code generation; depicted in this [diagram](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/code-generation/#our-approach).
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: represented as CRD in [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/crd.go#L63) is a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/4fd4b72d1a40237285232f1b16c1d13de4f1220d/models/apis/s3/2006-03-01/api-2.json#L1).
* `API inference` | `inference` : the discovery or determination of the structure of API resources, including the fields of said resources and the relationship between resources in an API; details [here](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/api-inference/).
* `ackmodel`: the *output* of the API inferece stage of the `pipeline`; represented in code as [Model](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/model.go#L36).


## Problem
*code-generator's* accelerated growth is resulting in tech debt that degrades the experience for `platform` maintainers and contributors. The codebase is becoming difficult to read and extend given the quick addition of features to onboard more services. This doc summarizes and introduces proposals to eliminate significant tech debt and bring clarity to the `pipeline` so that development on `ACK platform` is as pleasant and clear as it is open.


## Solution

### Overview
The tech debt will be reduced after stratifying `ackgenconfig` into catgories and refactoring Go code generating logic. Code will be moved, modified, and new packages will be created in order to accomplish this effort. A new command, `infer-model`, will be added to the code-generator `pipeline` to make the overall flow clearer and more efficient. `infer-model` will resolve relations between `ackgenconfig` and `aws-sdk` and caches the "inference" data so that it can be consumed by downstream commands, `apis` and `controller`.

### Requirements
* Code is easy to read, intuitive, and extendable
* Code generation processes remain clear and transparent
* Code generation features and templates are usable for all use cases

## Approach

### Split `ackgenconfig` into 2 categories
Move configs to its own pkg, `pkg/config`, then split `ackgenconfig` into 2 categories:
  * `pkg/config/model.go`: configuration to handle `API inference`
  * `pkg/config/generate.go`: configuration to handle and direct code generation functions



![current-config-access](./images/current_config_access.png)
* `r` represents a `resource` or `crd`
* `SetResource` calls `r.GetOutputShape(op)` to retrieve the shape for a given operation
* After, `SetResource` calls `r.GetOutputWrapperFieldPath(op)`; it's basically a wrapper for accessing `ackgenconfig`
* Finally, `r.getWrapperOutputShape()` is called recursively and resolves which shape to return
* `r` is doing a lot of work and it isn't clear what is being resolved


---

![proposed-config-access](./images/proposed_config_access.png)
* `m` represents `ackmodel`
* `SetResource` calls `ackmodel` helpers now instead of `crd`
* Under the hood, `ackmodel` will access `ackgenconfig` category to fetch requested config values
* With `m` being the source of truth and `API inference` logic, it is safe to assume the output shape you get back takes all configs into account
* Also, `resource` is no longer bogged down with helpers unrelated to a `resource`


### Generator enhancements
The `code` package is responsible for generating Go code, but has become overloaded with other functionality resulting in the accumulation of tech debt. This tech debt will be addressed in a number of ways:
* consolidating common logic/data
* reducing scope of overloaded methods
* removing duplicate code
* generalizing areas with hard-coded use cases.

### New command `./ack-generate model`
`model` takes `aws-sdk` and *generator.yaml* as **input**, discovers between the 2, persists and caches the data as serialized JSON, then **outputs** `ackmodel` (default location: `~./cache/aws-controllers-k8s/ack-model.json`). Existing commands, `apis` and `controller`, will be downstream and take `ackmodel` as an input. This will immediately improve the `pipeline` by removing duplicated `inference` work being done in both commands and clean up generator implementations.

 With a new `model` command, the code generation can flow from a common `inference` source which also creates opportunity for future commands:

![proposed-gen](./images/proposed_gen.png)


## Design Proposals
The efforts described above do not necessarily depend on one another, but I recommend reviewing and implementing in the order below:
   * [`ackgenconfig` Categories](./inference.md)
   * [Generator Enhancements](./generator.md)
   * [`model` Command](./model_cmd.md)