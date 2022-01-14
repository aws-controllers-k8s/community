### Key Terms
* `pipeline`: the collection of multiple phases involved in code generation; depicted in this [diagram](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/code-generation/#our-approach)
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: represented as CRD in [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/crd.go#L63) is a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/4fd4b72d1a40237285232f1b16c1d13de4f1220d/models/apis/s3/2006-03-01/api-2.json#L1)
* `API inference` | `inference` : logic involving relations between `resource`, `shape`, `ackgenconfig`, and `aws-sdk`. Details [here](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/api-inference/)
* `ackmodel`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/model.go#L36) representation of ACK's view of the world; the source of truth for `aws-sdk`, `ackgenconfig`, and `API inference`


## Problem
*code-generator's* accelerated growth is resulting in tech debt that degrades the experience for platform maintainers and contributors. The codebase is becoming difficult to read and extend given the quick addition of features to onboard more services. This doc summarizes and introduces proposals to eliminate significant tech debt and bring clarity to the `pipeline` so that development on ACK Platform is as pleasant and clear as it is open.


## Solution

### Overview
The tech debt will be reduced after stratifying `ackgenconfig` into catgories and refactoring Go code generating logic. Code will be moved, modified, and new packages will be created in order to accomplish this effort. A new command, `infer-model`, will be added to the code-generator `pipeline` to make the overall flow clearer and more efficient. `infer-model` will resolve relations between `ackgenconfig` and `aws-sdk` and caches the "inference" data so that it can be consumed by downstream commands, `apis` and `controller`.

### Requirements
* Code is easy to read, intuitive, and extendable
* Code generation processes remain clear and transparent
* Code generation features and templates are useable for all use cases

## Approach

### Split `ackgenconfig` into 2 categories
Move configs to its own pkg, `pkg/config`, then split `ackgenconfig` into 2 categories:
  * **inference**: configs used to *infer* a relation between `resource` and `aws-sdk` such as `renames` located in `pkg/generate/config/inference.go`
  * **code-generating**: configs used to instruct the code-generator on how to *generate* Go code for a resource such as `output_wrapper_field` located in `pkg/generate/config/generate.go`



![current-config-access](./images/current_config_access.png)
* `r` represents a `resource` or `crd`
* `SetResource` calls `r.GetOutputShape(op)` to retrieve the shape for a given operation
* Then, `r.GetOutputWrapperFieldPath(op)` is invoked; it's basically a wrapper for accessing `ackgenconfig`
* Finally, `r.getWrapperOutputShape()` is called recursively and resolves which shape to return
* `r` is doing a lot of work and it isn't clear what is being resolved


---

# TODO!!
![proposed-config-access](./images/proposed_config_access.png)
* `m` represents `ackmodel`
* `SetResource` calls `m.GetOutputShape(op)` now instead of `crd`
* Under the hood, `ackmodel` will access its `ackgenconfig` to get config values
* Then use helpers to resolve between `sdk-shape` and `ackgenconfig`
* With `m` being the source of truth and `API inference` logic, it is safe to assume the output shape you get back takes all configs into account
* Also, `resource` is no longer bogged down with helpers unrelated to a `resource`


### Generator enhancements
The `code` package is responsible for generating Go code, but has become overloaded with other functionality resulting in the accumulation of tech debt. This tech debt will be addressed in a number of ways:
* consolidating common logic/data
* reducing scope of overloaded methods
* removing duplicate code
* generalizing areas with hard-coded use cases.

# TODO: cleaner/better code diagram? Before/After


### New command `./ack-generate infer-model`
`infer-model` takes `aws-sdk` and *generator.yaml* as **input**, resolves relations/conflicts between the 2, persists and caches the data as serialized JSON, then **outputs** the `inferred-model` (default location: `~./cache/aws-controllers-k8s/ack-inferred-model.json`). Existing commands, `apis` and `controller`, will be downstream and take `inferred-model` as an input. This will immediately improve the `pipeline` by removing duplicated `inference` work being done in both commands and clean up generator implementations.

 With a new `infer-model` command, the code generation can flow from a common `inference` source which also creates opportunity for future commands:

![proposed-gen](./images/proposed_gen.png)


## Design Proposals
The efforts described above do not necessarily depend on one another, but I recommend reviewing and implementing in the order below:
   * [`ackgenconfig` Categories](./inference.md)
   * [Generator Enhancements](./generator.md)
   * [`infer-model` Command](./model_cmd_2.md)