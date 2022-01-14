# Generator enhancements

### Key Terms
* `ackgenconfig`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/config/config.go#L24) representation of *generator.yaml*. an **input** to *code-generator*.
* `resource` | `k8s-resource` | `ackcrd`: represented as CRD in [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/crd.go#L63) is a single top-level resource in an AWS Service. *code-generator* generates these resources using heuristics and `ackgenconfig`.
* `shape` | `aws-sdk` | `sdk-shape` | `sdk`: the original operations, models, errors, structs for a given AWS service. sourced from *aws-sdk*, ex: [aws-sdk-go s3](https://github.com/aws/aws-sdk-go/blob/4fd4b72d1a40237285232f1b16c1d13de4f1220d/models/apis/s3/2006-03-01/api-2.json#L1)
* `API inference` | `inference`: logic involving relations between `resource`, `shape`, `ackgenconfig`, and `aws-sdk`. Details [here](https://aws-controllers-k8s.github.io/community/docs/contributor-docs/api-inference/)
* `ackmodel`: the [code](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/model.go#L36) representation of ACK's view of the world; the source of truth for `aws-sdk`, `ackgenconfig`, and `API inference`

## Problem
The `code` package, the package responsible for **generating** Go code, is overloaded with unrelated functionality and tightly coupled to ACK's code-generating use case:
  * [overloaded functions](#overloaded-functions)
  * [hard-coded ACK use case](#ACK-use-case)
  * [unclear API access](#API-access)


## Solution
The solution is to refactor `pkg/code` as follows:
* **reduce the size of generator functions** (i.e. `SetResource`) by moving functionality unrelated to Go-code generation to new/more relevant package and by iterating over CRD Fields instead of an operation's input/output shape.
* **clarify generator APIs** by exposing public, general-use `Setters` (i.e. `SetResourceForContainer`) only
* **generalize generator code and templates** by replacing hard-coded use case (ACK, Crossplane) with `ackgenconfig` provided by the user

### Implementation

#### Move code unrelated to generation
Code unrelated to generating Go code will be moved to new and/or existing packages. This includes:
* private helpers such as [getSortedLateInitFieldsAndConfig](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/late_initialize.go#L55) in `late_initialize.go` move to `config` pkg
* public helpers such as [FindIdentifiersInShape](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/common.go#L33) like in `common.go` move to `model` pkg
* code within `SetResource` functions [directly accessing config and inferring](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/set_resource.go#L213-L253) like in `set_resource.go` move to `config` pkg for access and `model` pkg for inference logic.


#### Iterate over Fields instead of Operation.Members
:warning: **pre-requisites: [consolidate `inference` logic](./inference.md) & add data to `ackmodel` to persist inferences** :warning:
* This effort will clean up overloaded functions, `SetResource` & `SetSDK`, significantly by eliminating the need to fetch config and resolve within Go-code-generating logic. Instead of resolving a CRD's field with a field from an [operation shape](https://github.com/aws-controllers-k8s/code-generator/blob/b24c062600f1ae90d62e760c23e69651ac167a24/pkg/generate/code/set_resource.go#L142), the code-generating functions will iterate over the CRD's fields (i.e. treating provided Fields as source of truth) and generate necessary code.
```
// Current
for memberIndex, memberName := range outputShape.MemberNames() {...}

// Proposed
for fieldIndex, fieldName := range r.GetFields() {...}

```


#### Align access and interface for generator funcs
The goal is to make it easier for clients to use `SetResource` and `SetSDK` by removing the need to know the target's type.

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

#### Generalize areas with ACK-specific code
Today's code-generator use cases, ACK and Crossplane, should be able to leverage the same code/features and templates in code-generator.
* Features with ACK-specific code such as [late_initialize](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/late_initialize.go#L175) should check relevant config for prefixes/substitutions
* Remove specialized handling for ACK-specific use case such as [ACKResourceMetadata handling in SetResource](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/set_resource.go#L147-L182). Then, refactor CRD to include `ACKResourceMetadata` as a field in its Status, if defined in *generator.yaml*. Finally, after completing refactor to make the code generation **field-focused**, the `ACKResourceMetadata` will be processed like any other field and no specialized handling will be required.
* Remove hard-coded use cases (ACK, Crossplane) such as `ACKResourceMetadata` and `Conditions` from templates like [crd.go.tpl](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/templates/apis/crd.go.tpl#L30-L46). Not only does this make the template more general, but it also aligns with **field-focused** approach.
* After removing hard-coded use cases, expose new `ackgenconfig` so users can add code/directives for their use in *generator.yaml*. *Conceptual Render of ACK Use Case in generator.yaml:*
  
  ```
  # adds 'ACKResourceMetadata' and 'Conditions' fields to every CRD Status
      CRD:
        Fields:
          Status:
            - ACKResourceMetadata
            - Conditions

  ```



## Appendix

### Overloaded functions
* Overloaded functions in `code` pkg such as [SetResource](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L75) is responsible for a lot more than generating Go code:
  * [resolves config and shapes](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L120-L135)
  * [gets SetterConfig and resolves](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/generate/code/set_resource.go#L213-L253)


### ACK use case
There's ACK-specific code being hard-coded in functions relating to Go-code generation. As a result, other use cases such as *Crossplane* cannot leverage these functions despite most of the logic being the same. Examples:
* [ackerrors](https://github.com/aws-controllers-k8s/code-generator/blob/c7b19a3ec651b287477e7330d0ea1c725a904310/pkg/generate/code/set_resource.go#L706)
* [ACKResourceMetadata](https://github.com/aws-controllers-k8s/code-generator/blob/c7b19a3ec651b287477e7330d0ea1c725a904310/pkg/generate/code/set_resource.go#L928)
* [LateInitializeFromReadOne](https://github.com/aws-controllers-k8s/code-generator/blob/c7b19a3ec651b287477e7330d0ea1c725a904310/pkg/generate/code/late_initialize.go#L207)

Templates with hard-coded data:
* [crd.go.tpl](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/templates/apis/crd.go.tpl#L30-L46)


### API access
APIs exposed in `code` package such as `SetResource` and `SetSDK` allow for public access on *specific* `Setters` (i.e. [SetResourceForStruct](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/set_resource.go#L1217)), but not the *general* use case 
(i.e. [setResourceForContainer](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/generate/code/set_resource.go#L1154)). As a result, clients need to know implementation details such as which data type is being set in order to use the API.

