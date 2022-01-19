# Adding `infer-model` command

## Problem
There are inefficiencies in the code-generator `pipeline` that are starting to muddy the code and overall flow:

* Duplicate work between `api` and `controller` commands; both generate `ackmodel`-- specifically `inference` logic
* Confusing CX. There's a not-so-clear dependency on a folder in the $SERVICE-controller via [getLatestAPIVersion](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/cmd/ack-generate/command/common.go#L271). This goes against the flow of the `pipeline` and creates confusion as there is a `--version` flag for the `apis` command (the above function takes precedence).
* How *code-generator* resolves `ackgenconfig` with `aws-sdk` to create `ackmodel` is inconsistent (and some duplicated work). For example, overrides such as `custom_shapes` [edit aws-sdk directly](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/sdk/custom_shapes.go#L62-L63) while others such as `ignored_operations` [set operation values to nil](https://github.com/aws-controllers-k8s/code-generator/blob/26e5da2e7656bb836ee438c05df14f2adc50197d/pkg/model/model.go#L295) in memory leaving the actual operations in `aws-sdk` untouched.

Current Pipeline
---

![current-pipeline](./images/current_pipeline.png)
* Displayed are the important and mostly duplicated calls between the `apis` and `controller` commands
* The cache is only used to store the `aws-sdk` repo
* The red line denotes a dependency on $SERVICE-controller (output directory) repo that should be removed
* `GetCRDs` is duplicated between the commands when creating `ackmodel`

## Solution
Add a new command, `infer-model`, which takes `aws-sdk` and *generator.yaml* as **input**, resolves relations/conflicts between the 2, persists and caches the data as serialized JSON, then **outputs** the `inferred-model` (default location: `~./cache/aws-controllers-k8s/ack-inferred-model.json`). Existing commands, `apis` and `controller`, will be downstream and take `inferred-model` as an input-- no longer needing to execute `inference` logic themselves. Within code, the `inferred-model` will be a new attribute withinin `ackmodel`.


Updated Pipeline
---
![updated-pipeline](./images/proposed_pipeline.png)
* This is not an exhaustive diagram of the calls, but shows the clear responsibility of `generateModel` and how downstream commands like `api` and `controller` become significantly lighter and easier to follow.
* `inferred-model` will be cached in the same folder as `aws-sdk` and downstream commands will check for `inferred-model` in the cache when hydrating `ackmodel`
* No more dependency on $SERVICE-controller repo. The `ackmodel` will be generated with a specific version which can be extracted after commands `GetACKModel()`

Updated `ackmodel`:
---
Data relating to `API inference` will be moved from a top-level attribute in `ackmodel` to the `inferred-model` struct:

```
// serialized & cached during ./ack-generate infer-model
type InferredModel struct {
	apiVersion         string
	crds               []*CRD
	typeDefs           []*TypeDef
	typeImports        map[string]string
	typeRenames        map[string]string
}

...

type Model struct {
	SDKAPI             *SDKAPI
	servicePackageName string
    InferredModel    *InferredModel `json:"inferred_model"`
	// Instructions to the code generator how to handle the API and its
	// resources
	cfg *ackgenconfig.Config
}
```

`inferred-model` will be accessed through general-use `Model` helper functions. Ex:
```
// GetShapeRef returns the ShapeRef for a given resource and fieldName, rename-inclusive
func (m *Model) GetShapeRef(resource *CRD, fieldName string) *awssdkmodel.ShapeRef {}
```

### Requirements
* Code generation `pipeline` flows in a single direction
* Optimize `pipeline` by removing repeated work
* Logic to create `inferred-model`, parse `ackgenconfig`, and resolve differences with `aws-sdk` is centralized and consistent

## Implementation

### Prerequisites
* Work to achieve a **field-focused** approach covered in [`ackgenconfig` Categories proposal](./inference.md)
* Analyze and fill test gaps
* Any issues with marshalling `inferred-model` with proposed structs?

### Add `InferredModel` to `ackmodel`
Update `ackmodel` to include `InferredModel` as an attribute per the description above.

Fields may need to be altered/added to persist `API inference` data. For example, ensure the hydrating of [Field.ShapeRef](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/field.go#L48) reflects the Field's rename, if applicable. This will align with the **field-focused** model and eliminates the need for [SetResource](https://github.com/aws-controllers-k8s/code-generator/blob/main/pkg/generate/code/set_resource.go#L191-L208) to do this work.

Add/modify helper `ackmodel` methods to provide general use/access (not direct access) to `InferredModel` data like `m.GetShapeRef()` mentioned above. Downstream funcs like `SetResource` will become leaner after abstracting [`inference` work like rename checks](https://github.com/aws-controllers-k8s/code-generator/blob/d9d3390a4d5d39ccd4cab4fbdb5cef356211b01a/pkg/generate/code/set_resource.go#L185-L209) to `ackmodel` helpers.

### Create new `infer-model` command
* Define `cmd/ack-generate/infer-model.go`
* Move `api inference` logic like [CRDs creation](https://github.com/aws-controllers-k8s/code-generator/blob/82c294c2e8fc6ba23baa0034520e84351bb7a32f/pkg/model/model.go#L77) to `infer-model` command
* Marhsall & cache `inferred-model` in `--cache-dir`

#### Update `api` and `controller` commands
* Upon creating `ackmodel`, check cache for `inferred-model`; if none, exit and ask to re-run
* update calls in `build-controller` to pass in `inferred-model` parameter and execute `infer-model` command prior to `apis` and `controller`

#### In Scope
* add `infer-model` command
* update `pipeline` to use `infer-model` and pass artifacts downstream to `api` and `controller`
* make naming consistent on affected code
* tests related to this code
  * for a given `model` expect this generated code/types maybe?
* documentation updates for `pipeline` and codebase

#### Out of Scope
* changes to *generator.yaml* interface

### Test plan
* After relocating logic to `inferred-model`:
  * execute code-generator against X services
  * run the e2e test suite for each service
  * resolve issues
* Save off the marshalled `inferred-model` from above test into `testdata/`
* Update tests to generate `inferred-model`, then compare it to expected blob in `testdata/`


### Alternative Solutions

#### Odd lex, but okay
* write our own [Go interpreter](https://interpreterbook.com/) to take in `aws-sdk-go` and flags/config to generate controller code