# Centralize config access and reconciliation

### Key Terms
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/model/crd.go#L63) reprensenting a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/main/models/apis/s3/2006-03-01/api-2.json#L1)
* `reconciliation`: logic involving access to or relations between `resource`, `shape`, `ackgenconfig`, and `aws-sdk`
* `ackmodel`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/model/model.go#L36) representation of ACK's view of the world; the source of truth for `aws-sdk`, `ackgenconfig`, and `reconciliation`

## Problem
It is becoming increasingly difficult to contribute to and maintain *code-generator* due to unclear encapsulations and dispersed logic throughout the codebase. For example, `ackgenconfig` is passed throughout code and accessed in a variety of ways, and logic involving `reconciliation` is inconsistent, both in implementation and interface. Some specific examples include:
  * [overextended encapsulations](#overextended-encapsulations)
  * [long function parameter lists](#long-parameters)
  * [inconsistent use of `ackgenconfig`](#ackgenconfig-use)
  * [increased cognitive load for contributors](#more-work-for-contributors)


## Solution
The solution is to repair and realign encapsulations in `code-generator/pkg/`:
  * enforce `ackgenconfig` is **only** accessible via *Getters* exposed in [config package](https://github.com/aws-controllers-k8s/code-generator/tree/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/config).  
  * centralize `reconciliation` by extending `ackmodel` in [model.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/model/model.go#L36) 
  * update accessibility of *Setters* so only the relevant interfaces are public
    * ex: [SetResourceForStruct](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/set_resource.go#L1217) vs. [setResourceForContainer](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/set_resource.go#L1154)


### Requirements
* No impact to existing functionality
* Code is easy to read, intuitive, and extendable
* Naming in code is consistent (ex: Find vs. Get)
* Navigating code is easier because pkg responsibility is clear


### Prerequisites
*code-generator* contains unit tests and also runs a subset of $SERVICE-controller e2e tests against PRs. Are these comprehensive enough to detect breakage in the code that we'll be touching/refactoring? If there are gaps between tests and code being refactored, then **tests will need to be created or updated** prior to implementing the solution.

Agree on terms for consistent code. Terms such as Get vs. Find, CRDS/Resource, etc.

### Implementation

#### Consolidate `ackgenconfig` access and expose Getters
The purpose is to realign responsibilities of `CRD` and `Config` making code more consistent and intuitive.

1. Remove `ackgenconfig` as an attribute in `type CRD struct` in [crd.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/model/crd.go#L65) and fix breakage
  * move the associated methods to `config` package and update callers, ex: [GetOutputWrapperFieldPath](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L420):

```
// Current
wrapperFieldPath := r.GetOutputWrapperFieldPath(op)

// Proposed
wrapperFieldPath := cfg.GetOutputWrapperFieldPath(op)
```

2. Align `Get` naming and return types in `ackgenconfig` (i.e. Get__ , not Find__)
  * open to discussion on the exact wording/format as long as it's consistent. Proposing prepending funcs with `Get` and return the desired value and error to start

```
// Current
func (c *Config) ResourceConfig(name string) (*ResourceConfig, bool)

// Proposed
func (c *Config) GetResourceConfig(name string) (*ResourceConfig, error)

```

3. Look for other wrappers and access to `ackgenconfig` and consildate to `config` package

```
// Current
func getSortedLateInitFieldsAndConfig(cfg *ackgenconfig.Config, r *model.CRD) ([]string, map[string]*ackgenconfig.LateInitializeConfig)


// Proposed
func (c *Config) GetLateInitFieldsAndConfig(cfg *ackgenconfig.Config, r *model.CRD) (map[string]*ackgenconfig.LateInitializeConfig, error)

// Note, this simply gets the config; sorting and any processing AFTER getting config is done by the client

```

#### Centralize `reconciliation`
Similar to the above, move `reconcialiation`-related functions from `code` package to `ackmodel` to realign responsibilities.

1. Refactor `type CRD struct` in [crd.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/model/crd.go#L64) to remove `sdkapi`  as attribute and fix breakage
  * move methods associated with attributes to `ackmodel`

```
// Current
func (r *CRD) TypeRenames() map[string]string

//Proposed
func (m *Model) GetTypeRenames() (map[string]string, error)

```

2. Move code in `pkg/code/` related to `reconciliation` to `pkg/model/`
  * previous dependencies on invalid helpers (ex: `r.GetAllRenames`) should be resolved
  * ex: most/all of [common.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/common.go) has nothing to do with generating code; it retrieves information related to `aws-sdk` and `ackgenconfig` regarding `shapes`, `ops`

```
// Current
func FindIdentifiersInShape(r *model.CRD, shape *awssdkmodel.Shape, op *awssdkmodel.Operation) []string

// Proposed
func (m *Model) GetIdentifiersInShape(r *model.CRD, shape *awssdkmodel.Shape, op *awssdkmodel.Operation) ([]string, error)

```

* Note, for scenarios that do both `reconcile` and output Go code (ex: [late_initialize::FindLateInitializedFieldNames](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/late_initialize.go#L31), these will need to be refactored into separate pieces prior to migration


#### Align access/interface for `__ForStruct` funcs
The goal is to make it easier for clients to use `SetResource` and `SetSDK` by removing the need to know the target type.

1. Refactor access (public vs. private) for `SetResourceFor___`, and `SetSDKFor___`
  * `setResourceForContainer` and `setSDKForContainer` are private while `SetResourceForStruct` and `SetSDKForStruct` are public despite the former pair being more general than the latter. Access should be swapped so users calling this in [controller.go](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/ack/controller.go#L102) don't need to worry about Struct or what type.. just pass container. Leaves it open-ended for override types too.
  * `GoCodeCompare` vs `GoCodeCompareStruct` .. can [these](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/ack/controller.go#L119-L122) be consolidated?

```
// Current
SetResourceForStruct()
setResourceForContainer()
setResourceForSlice()
setResourceForMap()
setResourceForScalar()

// access in controller.go
"GoCodeSetResourceForStruct" : ... return code.SetResourceForStruct()

// Proposed
SetResourceForContainer()
setResourceForStruct()
setResourceForSlice()
setResourceForMap()
setResourceForScalar()

// access in controller.go
"GoCodeSetResourceForContainer" : ... return code.SetResourceForContainer()

```


#### In Scope
* Centralize code related to `reconciliation` to `model` package
* Expose Getters in `config` and replace previous helpers/wrappers
* Update public/private interfaces so users need only focus on the General case and not be bogged down with config details
* Align naming and return types for affected code
* Consolidate `SetterConfig` access/logic to `ackgenconfig` and/or `ackmodel`


#### Out of Scope
* Changing logic/algorithms
  * the goal is minimal impact to functionality; once everything is in its proper place we can address inefficiencies separately
* Consistent styling/naming throughout codebase

### Test plan
* add "sufficient" testing beforehand for quick feedback
* update existing unit tests to reflect any changed logic due to migration
* run a "healthy" subset of service controller e2e tests for validation

## Appendix

### Long paramaters
The majority of functions in `code` package take both `ackgenconfig` and `ackresource` as parameters, but this is unnecessary because [`ackresource` has `ackgenconfig` as an attribute](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/model/crd.go#L65):
  * [CheckExceptionMessage](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/check.go#L36)
  * [CompareResource](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/compare.go#L77)
  * [SetResource](https://github.com/aws-controllers-k8s/code-generator/blob/f37b8ea9fcc401d66018d7849e8c809da5b4c99c/pkg/generate/code/set_resource.go#L74)
  * [SetSDK](https://github.com/aws-controllers-k8s/code-generator/blob/f37b8ea9fcc401d66018d7849e8c809da5b4c99c/pkg/generate/code/set_sdk.go#L75)


### ackgenconfig use
`ackgenconfig` is acessed in various layers & ways:
  * [creating `ackmodel`](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/cmd/ack-generate/command/common.go#L231)
  * [reconciling resource/shape logic](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/common.go#L54)
  * [generating Go code](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/set_resource.go#L190)
    * Note the lines invoking `cfg.ResourceFieldRename` and `r.HasMember`. The former is direct access while the latter is indirect access as it calls [r.Config().ResourceFieldRename()](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L747) under the hood. There are many instances of this happening throughout code because [CRD struct](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L63) has `ackgenconfig` as an attribute.

### Overextended encapsulations
* The [code package](https://github.com/aws-controllers-k8s/code-generator/tree/main/pkg/generate/code) is responsible for generating a controller's *Go* code, but it overextends itself in multiple areas by also trying to reconcile resource/aws-shape logic:
  * [common::FindPluralizedIdentifiersInShape()](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/common.go#L97)
  * [set_resource::setResourceReadMany()](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/set_resource.go#L466-L479)
  * [late_initialize::getSortedLateInitFieldsAndConfig()](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/late_initialize.go#L55)

* `CRD struct` should have neither `sdkAPI` nor `cfg` as attributes because it represents a [single top-level resource](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/crd.go#L62) -- there's no need for it to have all knowledge of the ACK universe. As a result, `ackresource` has multiple helpers unrelated to the resource itself:
  * [GetOutputShape](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L442)
  * [GetAllRenames](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L644)
  * [GetOutputWrapperFieldPath](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/crd.go#L420)

* Helpers like [FindIdentifiersInShape](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/common.go#L31-L33) are implemented in the `code` package, but are not related to code generation


### More work for contributors
Developers needs to keep a mental model of which configs have been applied or need to be consulted:
  * [//TODO: should these fields be renamed before looking them up in spec?](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/set_resource.go#L143)
  * [//TODO: check generator config for exceptions?](https://github.com/aws-controllers-k8s/code-generator/blob/25c43e827527b43c652b6e1995265c8d6027567f/pkg/generate/code/set_sdk.go#L231)

And because adding new features requires consulting config throughout code, some "lower priority" areaas get put on the back burner:
  * [//TODO: should we support overriding these fields?](https://github.com/aws-controllers-k8s/code-generator/blob/59c6892e4b61b5e2076e5e5504daba8278b82980/pkg/model/model.go#L219)