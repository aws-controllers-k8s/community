# `ackgenconfig` categories

### Key Terms
* `pipeline`: the collection of multiple phases involved in code generation; depicted in this [diagram](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/code-generation/#our-approach)
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: represented as CRD in [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/crd.go#L63) is a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/4fd4b72d1a40237285232f1b16c1d13de4f1220d/models/apis/s3/2006-03-01/api-2.json#L1)
* `API inference` | `inference` : logic involving relations between `resource`, `shape`, `ackgenconfig`, and `aws-sdk`. Details [here](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/api-inference/)
* `ackmodel`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/model.go#L36) representation of ACK's view of the world; the source of truth for `aws-sdk`, `ackgenconfig`, and `API inference`

## Problem
It is becoming increasingly difficult to contribute to and maintain *code-generator* due to unclear encapsulations and dispersed logic throughout the codebase. For example, `ackgenconfig` is passed throughout being accessed in a variety of ways, and logic involving `API inference` is inconsistent, both in implementation and interface. Specific examples include:
  * [overextended encapsulations](#overextended-encapsulations)
  * [long function parameter lists](#long-parameters)
  * [inconsistent use of `ackgenconfig`](#ackgenconfig-use)
  * [increased cognitive load for contributors](#more-work-for-contributors)


## Solution
The proposal is to delineate `ackgenconfig` into 2 categories, **inference** and **code-generating**. The structures, `Getters/Setters`, and helpers will be defined in the `config` package (migrate from `pkg/generate/config` to `pkg/config`) and encapsulated in files by category:
  * **inference**: configs used to *infer* a relation between `resource` and `aws-sdk` such as `renames`
  * **code-generating**: configs used to instruct the code-generator on how to *generate* Go code for a resource such as `output_wrapper_field`

Encapsulation will be improved because functions consuming `ackgenconfig` will grab configurations *relevant to its responsibility only*. For example, functions responsible for generating Go code, ex: `SetResource`, need only **code-generating** configs; therefore, one can eliminate any calls/logic relating to **inference** or other `ackgenconfig` parsing.

TODO: This new classification makes navigating the code more intuitive and consistent as well. Instead of using helpers in `model/crd.go` or `generate/common.go` to read config/resolve an inference, clients will know to use and extend `config` package. It will also be easier to enfore consistent style and interfaces after consolidating into the same location, `config` package.

### Requirements
* Code is easy to read, intuitive, and extendable
* Naming in code is consistent (ex: Find vs. Get)
* Navigating code is easier because pkg responsibility is clear
* `ackgenconfig` uses a **field-focused** model


### Prerequisites
*code-generator* contains unit tests and also runs a subset of $SERVICE-controller e2e tests against PRs. We need to identify the affected code paths and determine whether these are being covered in existing tests. If there are gaps, then **tests will need to be created or updated** prior to implementing the solution.

Agree on terms for consistent code. Terms such as Get vs. Find, CRDS/Resource, etc.

### Implementation

#### Consolidate `ackgenconfig` access and expose Getters
The purpose is to realign responsibilities of `CRD` and `Config` making code more consistent and intuitive.

1. Move `ackgenconfig` to its own package, `pkg/config`

2. Remove `ackgenconfig` as an attribute in `type CRD struct` in [crd.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/model/crd.go#L65) and fix breakage
  * move the associated methods to `config` package and update callers, ex: [GetOutputWrapperFieldPath](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L420):

```
// Current
wrapperFieldPath := r.GetOutputWrapperFieldPath(op)

// Proposed
wrapperFieldPath := cfg.GetOutputWrapperFieldPath(op)
```

3. Align `Get` naming and return types in `ackgenconfig` (i.e. Get__ , not Find__)
  * open to discussion on the exact wording/format as long as it's consistent. Proposing prepending funcs with `Get` and return the desired value and error to start

```
// Current
func (c *Config) ResourceConfig(name string) (*ResourceConfig, bool)

// Proposed
func (c *Config) GetResourceConfig(name string) (*ResourceConfig, error)

```

4. Look for other wrappers and access to `ackgenconfig` and consildate to `config` package

```
// Current
func getSortedLateInitFieldsAndConfig(cfg *ackgenconfig.Config, r *model.CRD) ([]string, map[string]*ackgenconfig.LateInitializeConfig)


// Proposed
func (c *Config) GetLateInitFieldsAndConfig(cfg *ackgenconfig.Config, r *model.CRD) (map[string]*ackgenconfig.LateInitializeConfig, error)

Note, `Getters` will contain some "helper" logic such as sanitizing user-input and sorting, when applicable. This is safe because user-input (via `generator.yaml`) should never need to preserve order.

```
#### Define `ackgenconfig` Categories
After centralizing all config data and functions to `config` pkg, break `ackgenconfig` into **inference** and **code-generating** categories where each will encapsulate data and methods (Getters, Setters, helpers) in their respective files:
* `pkg/generate/config/inference.go`
* `pkg/generate/config/generate.go`

```
TODO: Config struct after these new fields are added
note don't want to expose this detail in interface so need to convert/hydrate internally
```

Next, update `ackgenconfig` accessors throughout code:

```
TODO: Show SetResource using config category config.GenerateConfig.UnwrapOrDefault()
still resolving inference on its own
```

#### Centralize `API Inference` helpers to `ackmodel`
Similar to the above, move `API inference`-related functions to `ackmodel` to realign responsibilities.

1. Refactor `type CRD struct` in [crd.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/model/crd.go#L64) to remove `sdkapi`  as attribute and fix breakage
  * move methods associated with attributes to `ackmodel`

```
// Current
func (r *CRD) TypeRenames() map[string]string

//Proposed
func (m *Model) GetTypeRenames() (map[string]string, error)

```

2. Move code in `pkg/code/` related to `API inference` to `pkg/model/`
  * previous dependencies on invalid helpers (ex: `r.GetAllRenames`) should be resolved
  * ex: most/all of [common.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/common.go) has nothing to do with generating code; it retrieves information related to `aws-sdk` and `ackgenconfig` regarding `shapes`, `ops`
  * update callers

```
// Current
func FindIdentifiersInShape(r *model.CRD, shape *awssdkmodel.Shape, op *awssdkmodel.Operation) []string

// Proposed
func (m *Model) GetIdentifiersInShape(r *model.CRD, shape *awssdkmodel.Shape, op *awssdkmodel.Operation) ([]string, error)

```

* Note, for scenarios that do both `API inference` and output Go code (ex: [late_initialize::FindLateInitializedFieldNames](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/late_initialize.go#L31), these will need to be refactored into separate pieces prior to migration


#### Field-focused `ackgenconfig`
Exposing `ackgenconfigs` in operation config (i.e. [renames](https://github.com/aws-controllers-k8s/code-generator/blob/c7b19a3ec651b287477e7330d0ea1c725a904310/pkg/generate/config/resource.go#L237-L240)) is an unclear experience for both users of the [interface](https://github.com/aws-controllers-k8s/s3-controller/blob/d8b7ab6c4d9a162f9736a3221a680f63028ba757/generator.yaml#L103-L110) and maintainers of the [implementation](https://github.com/aws-controllers-k8s/code-generator/blob/79311e52117a2df3f99eb921d1aa19c373bd6c7e/pkg/model/crd.go#L644). Moving to a **field-focused** is a more intuitive experience:
  * **Updated interface** no longer requires user to know which operations use this field
  ```
  resources:
    Bucket:
      fields:
        Name:
          <RENAME_CONFIG> #sources value from Bucket

  ```
  * **Updated implementation** has clearer encapsulation compared to `r.GetAllRenames(op)`, a `CRD` requiring an `op` to find `Field` renames.

  ```
  // GetRenames returns all renames for the provided field defined in the generator config
  func (m *Model) GetRenames(field *Field) []string {
    ...
    return field.Renames()
  }
  ```

  1. Update [FieldConfig](https://github.com/aws-controllers-k8s/code-generator/blob/b24c062600f1ae90d62e760c23e69651ac167a24/pkg/generate/config/field.go#L298) to support migrated config (i.e. RenamesConfig). Align structure with agreed upon interface.
  
  2. Update parsing logic to route new configs in `FieldConfig` to corresponding `ackgenconfig` attributes

  3. Add/modify Getters, Setters, and helpers for `FieldConfig`

  4. Refactor clients using op-based processing to **field-focused** approach

  ```
  SetResource() --> for _, _ range := CRD.Fields instead of op.shape.members
  - remove "Is this field in Spec/Status"?
  - For each field do a Model.GetShape(fieldName) 
    - //checks for "fieldName" in aws-sdk, if none found, check Fields.Renames[fieldName] or similar. returns corresponding shape in aws-sdk 

  how to replace targetAdaptedVarName += cfg.PrefixConfig.SpecField
?

  ```


#### In Scope
* Centralize code related to `API inference` to `model` package
* Expose Getters in `config` and replace previous helpers/wrappers
* Update public/private interfaces so users need only focus on the General case and not be bogged down with config details
* Align naming and return types
* Resolve 3 different `op` and `op` types used throughout code
* Migrate from operation-centric to field-centric approach for inference
* Change code structure by adding/moving packages for clearer encapsulations, ex: moving `pkg/generate/config` to `pkg/config`
* Changes to `generator.yaml` interface


#### Out of Scope
* *operation-centric* approaches outside of `ackgenconfig` such as [resource-identifier heuristic](https://github.com/aws-controllers-k8s/code-generator/blob/c7b19a3ec651b287477e7330d0ea1c725a904310/pkg/model/op.go#L48)


### Test plan
* add "sufficient" testing beforehand for quick feedback
* update existing unit tests to reflect any changed logic due to migration
* run a "healthy" subset of service controller e2e tests for validation

## Appendix

### Long parameters
The majority of functions in `code` package take both `ackgenconfig` and `ackresource` as parameters, but this is unnecessary because [`ackresource` has `ackgenconfig` as an attribute](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/crd.go#L65):
  * [CheckExceptionMessage](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/check.go#L36)
  * [CompareResource](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/compare.go#L77)
  * [SetResource](https://github.com/aws-controllers-k8s/code-generator/blob/f37b8ea9fcc401d66018d7849e8c809da5b4c99c/pkg/generate/code/set_resource.go#L74)
  * [SetSDK](https://github.com/aws-controllers-k8s/code-generator/blob/f37b8ea9fcc401d66018d7849e8c809da5b4c99c/pkg/generate/code/set_sdk.go#L75)


### ackgenconfig use
`ackgenconfig` accessed in various layers & ways:
  * [creating `ackmodel`](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/cmd/ack-generate/command/common.go#L231)
  * [API inference](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/common.go#L54)
  * [generating Go code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L190)
    * Note the lines invoking `cfg.ResourceFieldRename` and `r.HasMember`. The former is direct access while the latter is indirect access as it calls [r.Config().ResourceFieldRename()](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L747) under the hood. There are many instances of this happening throughout code because [CRD struct](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L63) has `ackgenconfig` as an attribute.

### Overextended encapsulations
* The [code package](https://github.com/aws-controllers-k8s/code-generator/tree/main/pkg/generate/code) is responsible for generating a controller's *Go* code, but it overextends itself in multiple areas by also implementing API inference logic:
  * [common::FindPluralizedIdentifiersInShape()](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/common.go#L97)
  * [set_resource::setResourceReadMany()](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L466-L479)
  * [late_initialize::getSortedLateInitFieldsAndConfig()](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/late_initialize.go#L55)

* `CRD struct` should have neither `sdkAPI` nor `cfg` as attributes because it represents a [single top-level resource](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L62) -- there's no need for it to have all knowledge of the ACK universe. As a result, `ackresource` has multiple helpers unrelated to the resource itself:
  * [GetOutputShape](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L442)
  * [GetAllRenames](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L644)
  * [GetOutputWrapperFieldPath](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L420)

* Helpers like [FindIdentifiersInShape](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/common.go#L31-L33) are implemented in the `code` package, but are not related to code generation


### More work for contributors
Developers need to keep a mental model of which configs have been applied or need to be consulted:
  * [//TODO: should these fields be renamed before looking them up in spec?](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L143)
  * [//TODO: check generator config for exceptions?](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/set_sdk.go#L231)

And because adding new features requires consulting config throughout code, some "lower priority" areas get put on the back burner:
  * [//TODO: should we support overriding these fields?](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/model.go#L219)
