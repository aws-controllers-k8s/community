# Understanding generator.yaml Configuration

This document describes the various configuration fields in a `generator.yaml` file that can be used to control the API inference and code generation for an ACK controller.

We will show examples of configuring specific ACK controllers to highlight various configuration options.

## Generate a resource manager package

For this section, we will use [ECR](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/Welcome.html) as our example service API.

When creating a new ACK controller, after running the [`controller-bootstrap`](https://github.com/aws-controllers-k8s/controller-bootstrap) program, you will be left with a `generator.yaml` that has all inferred API resources ignored.

For the ECR controller, the `generator.yaml` file would look like this:

```yaml=
ignore:
  resources:
    - Repository
    - PullThroughCacheRule
```

If we ran `make build-controller SERVICE=ecr` with the above `generator.yaml` file, we would have some basic directories and files created:

```bash=
[jaypipes@thelio code-generator]$ make build-controller SERVICE=ecr
building ack-generate ... ok.
==== building ecr-controller ====
Copying common custom resource definitions into ecr
Building Kubernetes API objects for ecr
Generating deepcopy code for ecr
Generating custom resource definitions for ecr
Building service controller for ecr
Generating RBAC manifests for ecr
Running gofmt against generated code for ecr
Updating additional GitHub repository maintenance files
==== building ecr-controller release artifacts ====
Building release artifacts for ecr-v0.0.0-non-release
Generating common custom resource definitions
Generating custom resource definitions for ecr
Generating RBAC manifests for ecr
```

```
[jaypipes@thelio ecr-controller]$ tree apis/ config/ pkg/
apis/
└── v1alpha1
    ├── ack-generate-metadata.yaml
    ├── doc.go
    ├── enums.go
    ├── generator.yaml
    ├── groupversion_info.go
    └── types.go
config/
├── controller
│   ├── deployment.yaml
│   ├── kustomization.yaml
│   └── service.yaml
├── crd
│   ├── common
│   │   ├── bases
│   │   │   ├── services.k8s.aws_adoptedresources.yaml
│   │   │   └── services.k8s.aws_fieldexports.yaml
│   │   └── kustomization.yaml
│   └── kustomization.yaml
├── default
│   └── kustomization.yaml
├── overlays
│   └── namespaced
│       ├── kustomization.yaml
│       ├── role-binding.json
│       └── role.json
└── rbac
    ├── cluster-role-binding.yaml
    ├── cluster-role-controller.yaml
    ├── kustomization.yaml
    ├── role-reader.yaml
    ├── role-writer.yaml
    └── service-account.yaml
pkg/
├── resource
│   └── registry.go
└── version
    └── version.go

11 directories, 25 files
```

To begin generating a particular resource manager, comment out the name of the resource from the ignore list and run `make build-controller SERVICE=$SERVICE`.

```yaml=
ignore:
  resources:
>   #- Repository
    - PullThroughCacheRule
```

After doing so, the resource manager for `Repository` resources will have been generated in the `ecr-controller` source code repository.

```
[jaypipes@thelio ecr-controller]$ tree apis/ config/ pkg/
apis/
└── v1alpha1
    ├── ack-generate-metadata.yaml
    ├── doc.go
    ├── enums.go
    ├── generator.yaml
    ├── groupversion_info.go
    ├── repository.go
    ├── types.go
    └── zz_generated.deepcopy.go
config/
├── controller
│   ├── deployment.yaml
│   ├── kustomization.yaml
│   └── service.yaml
├── crd
│   ├── bases
│   │   └── ecr.services.k8s.aws_repositories.yaml
│   ├── common
│   │   ├── bases
│   │   │   ├── services.k8s.aws_adoptedresources.yaml
│   │   │   └── services.k8s.aws_fieldexports.yaml
│   │   └── kustomization.yaml
│   └── kustomization.yaml
├── default
│   └── kustomization.yaml
├── overlays
│   └── namespaced
│       ├── kustomization.yaml
│       ├── role-binding.json
│       └── role.json
└── rbac
    ├── cluster-role-binding.yaml
    ├── cluster-role-controller.yaml
    ├── kustomization.yaml
    ├── role-reader.yaml
    ├── role-writer.yaml
    └── service-account.yaml
pkg/
├── resource
│   ├── registry.go
│   └── repository
│       ├── delta.go
│       ├── descriptor.go
│       ├── identifiers.go
│       ├── manager_factory.go
│       ├── manager.go
│       ├── references.go
│       ├── resource.go
│       ├── sdk.go
│       └── tags.go
└── version
    └── version.go

13 directories, 37 files
```

Note the new files under `pkg/resource/repository/`, `apis/v1alpha1/repository.go` and `config/crd/bases/`.  These files represent the Go type for the generated `Repository` custom resource definition (CRD), the resource manager package and the YAML representation for the CRD, respectively.

## Renaming things

Why might we want to rename fields or resources? Generally, there are two reasons for this:
 - reducing stutter in the input shape
 - correcting instances where a field is named differently in the input and output shapes

The first reason is to reduce "stutter" (or redundancy) in naming. For example, the ECR [`Repository`](https://github.com/aws-controllers-k8s/ecr-controller/blob/cefc45b65e1a560b9666a1544f26f955c6f84d36/generator.yaml#L2) resource has a field called `RepositoryName`.  This field is redundantly named because the resource itself is called `Repository`.  Every Kubernetes object has a `Metadata.Name` field and we like to align resource "name fields" with this simple `Name` moniker.

For this example, let's go ahead and "destutter" the `RepositoryName` field. To do this, we use the `renames` configuration option, specifying the input and output shapes and their members that we want to rename:

```yaml=
ignore:
  resources:
    #- Repository
    - PullThroughCacheRule
resources:
  Repository:
>   renames:
>     operations:
>       CreateRepository:
>         input_fields:
>           RepositoryName: Name
>       DeleteRepository:
>         input_fields:
>           RepositoryName: Name
>       DescribeRepositories:
>         input_fields:
>           RepositoryName: Name
```

> 📝 Note that we must tell the code generator which fields to rename in the input shapes for each API operation that the resource manager will call.  In the case of ECR `Repository` resources, the resource manager calls the `CreateRepository`, `DeleteRepository` and `DescribeRepositories` API calls and so we need specify the `RepositoryName` member field in each of those input shapes should be renamed to `Name`.

After calling `make build-controller SERVICE=ecr`, we see the above generator configuration items produced the following diff:

```diff=
diff --git a/apis/v1alpha1/ack-generate-metadata.yaml b/apis/v1alpha1/ack-generate-metadata.yaml
index e34e029..f214b43 100755
--- a/apis/v1alpha1/ack-generate-metadata.yaml
+++ b/apis/v1alpha1/ack-generate-metadata.yaml
@@ -1,13 +1,13 @@
 ack_generate_info:
-  build_date: "2022-11-09T20:15:42Z"
+  build_date: "2022-11-09T20:16:52Z"
   build_hash: 5ee0ac052c54f008dff50f6f5ebb73f2cf3a0bd7
   go_version: go1.18.1
   version: v0.20.1-4-g5ee0ac0
-api_directory_checksum: 0a514bef9cff983f9fe28f080d85725ccf578060
+api_directory_checksum: 84fb59a0991980da922a385f585111a1ff784d82
 api_version: v1alpha1
 aws_sdk_go_version: v1.44.93
 generator_config_info:
-  file_checksum: 87446926d73abae9355e6328eb7f8f668b16b18e
+  file_checksum: a383007f82a686dc544879792dde7b091aeededa
   original_file_name: generator.yaml
 last_modification:
   reason: API generation
diff --git a/apis/v1alpha1/generator.yaml b/apis/v1alpha1/generator.yaml
index cb7045a..ed0130f 100644
--- a/apis/v1alpha1/generator.yaml
+++ b/apis/v1alpha1/generator.yaml
@@ -2,3 +2,16 @@ ignore:
   resource_names:
     #- Repository
     - PullThroughCacheRule
+resources:
+  Repository:
+    renames:
+      operations:
+        CreateRepository:
+          input_fields:
+            RepositoryName: Name
+        DeleteRepository:
+          input_fields:
+            RepositoryName: Name
+        DescribeRepositories:
+          input_fields:
+            RepositoryName: Name
diff --git a/apis/v1alpha1/repository.go b/apis/v1alpha1/repository.go
index c226d4f..fc6165d 100644
--- a/apis/v1alpha1/repository.go
+++ b/apis/v1alpha1/repository.go
@@ -35,15 +35,15 @@ type RepositorySpec struct {
 	// be overwritten. If IMMUTABLE is specified, all image tags within the repository
 	// will be immutable which will prevent them from being overwritten.
 	ImageTagMutability *string `json:"imageTagMutability,omitempty"`
-	// The Amazon Web Services account ID associated with the registry to create
-	// the repository. If you do not specify a registry, the default registry is
-	// assumed.
-	RegistryID *string `json:"registryID,omitempty"`
 	// The name to use for the repository. The repository name may be specified
 	// on its own (such as nginx-web-app) or it can be prepended with a namespace
 	// to group the repository into a category (such as project-a/nginx-web-app).
 	// +kubebuilder:validation:Required
-	RepositoryName *string `json:"repositoryName"`
+	Name *string `json:"name"`
+	// The Amazon Web Services account ID associated with the registry to create
+	// the repository. If you do not specify a registry, the default registry is
+	// assumed.
+	RegistryID *string `json:"registryID,omitempty"`
 	// The metadata that you apply to the repository to help you categorize and
 	// organize them. Each tag consists of a key and an optional value, both of
 	// which you define. Tag keys can have a maximum character length of 128 characters,
diff --git a/apis/v1alpha1/zz_generated.deepcopy.go b/apis/v1alpha1/zz_generated.deepcopy.go
index 93919be..88dd4c0 100644
--- a/apis/v1alpha1/zz_generated.deepcopy.go
+++ b/apis/v1alpha1/zz_generated.deepcopy.go
@@ -421,13 +421,13 @@ func (in *RepositorySpec) DeepCopyInto(out *RepositorySpec) {
 		*out = new(string)
 		**out = **in
 	}
-	if in.RegistryID != nil {
-		in, out := &in.RegistryID, &out.RegistryID
+	if in.Name != nil {
+		in, out := &in.Name, &out.Name
 		*out = new(string)
 		**out = **in
 	}
-	if in.RepositoryName != nil {
-		in, out := &in.RepositoryName, &out.RepositoryName
+	if in.RegistryID != nil {
+		in, out := &in.RegistryID, &out.RegistryID
 		*out = new(string)
 		**out = **in
 	}
diff --git a/config/crd/bases/ecr.services.k8s.aws_repositories.yaml b/config/crd/bases/ecr.services.k8s.aws_repositories.yaml
index 438785e..9657569 100644
--- a/config/crd/bases/ecr.services.k8s.aws_repositories.yaml
+++ b/config/crd/bases/ecr.services.k8s.aws_repositories.yaml
@@ -61,17 +61,17 @@ spec:
                   all image tags within the repository will be immutable which will
                   prevent them from being overwritten.
                 type: string
-              registryID:
-                description: The Amazon Web Services account ID associated with the
-                  registry to create the repository. If you do not specify a registry,
-                  the default registry is assumed.
-                type: string
-              repositoryName:
+              name:
                 description: The name to use for the repository. The repository name
                   may be specified on its own (such as nginx-web-app) or it can be
                   prepended with a namespace to group the repository into a category
                   (such as project-a/nginx-web-app).
                 type: string
+              registryID:
+                description: The Amazon Web Services account ID associated with the
+                  registry to create the repository. If you do not specify a registry,
+                  the default registry is assumed.
+                type: string
               tags:
                 description: The metadata that you apply to the repository to help
                   you categorize and organize them. Each tag consists of a key and
@@ -92,7 +92,7 @@ spec:
                   type: object
                 type: array
             required:
-            - repositoryName
+            - name
             type: object
           status:
             description: RepositoryStatus defines the observed state of Repository
diff --git a/generator.yaml b/generator.yaml
index cb7045a..ed0130f 100644
--- a/generator.yaml
+++ b/generator.yaml
@@ -2,3 +2,16 @@ ignore:
   resource_names:
     #- Repository
     - PullThroughCacheRule
+resources:
+  Repository:
+    renames:
+      operations:
+        CreateRepository:
+          input_fields:
+            RepositoryName: Name
+        DeleteRepository:
+          input_fields:
+            RepositoryName: Name
+        DescribeRepositories:
+          input_fields:
+            RepositoryName: Name
diff --git a/helm/crds/ecr.services.k8s.aws_repositories.yaml b/helm/crds/ecr.services.k8s.aws_repositories.yaml
index 438785e..9657569 100644
--- a/helm/crds/ecr.services.k8s.aws_repositories.yaml
+++ b/helm/crds/ecr.services.k8s.aws_repositories.yaml
@@ -61,17 +61,17 @@ spec:
                   all image tags within the repository will be immutable which will
                   prevent them from being overwritten.
                 type: string
-              registryID:
-                description: The Amazon Web Services account ID associated with the
-                  registry to create the repository. If you do not specify a registry,
-                  the default registry is assumed.
-                type: string
-              repositoryName:
+              name:
                 description: The name to use for the repository. The repository name
                   may be specified on its own (such as nginx-web-app) or it can be
                   prepended with a namespace to group the repository into a category
                   (such as project-a/nginx-web-app).
                 type: string
+              registryID:
+                description: The Amazon Web Services account ID associated with the
+                  registry to create the repository. If you do not specify a registry,
+                  the default registry is assumed.
+                type: string
               tags:
                 description: The metadata that you apply to the repository to help
                   you categorize and organize them. Each tag consists of a key and
@@ -92,7 +92,7 @@ spec:
                   type: object
                 type: array
             required:
-            - repositoryName
+            - name
             type: object
           status:
             description: RepositoryStatus defines the observed state of Repository
diff --git a/pkg/resource/repository/delta.go b/pkg/resource/repository/delta.go
index a15d260..57b54df 100644
--- a/pkg/resource/repository/delta.go
+++ b/pkg/resource/repository/delta.go
@@ -77,6 +77,13 @@ func newResourceDelta(
 			delta.Add("Spec.ImageTagMutability", a.ko.Spec.ImageTagMutability, b.ko.Spec.ImageTagMutability)
 		}
 	}
+	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
+		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
+	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
+		if *a.ko.Spec.Name != *b.ko.Spec.Name {
+			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
+		}
+	}
 	if ackcompare.HasNilDifference(a.ko.Spec.RegistryID, b.ko.Spec.RegistryID) {
 		delta.Add("Spec.RegistryID", a.ko.Spec.RegistryID, b.ko.Spec.RegistryID)
 	} else if a.ko.Spec.RegistryID != nil && b.ko.Spec.RegistryID != nil {
@@ -84,13 +91,6 @@ func newResourceDelta(
 			delta.Add("Spec.RegistryID", a.ko.Spec.RegistryID, b.ko.Spec.RegistryID)
 		}
 	}
-	if ackcompare.HasNilDifference(a.ko.Spec.RepositoryName, b.ko.Spec.RepositoryName) {
-		delta.Add("Spec.RepositoryName", a.ko.Spec.RepositoryName, b.ko.Spec.RepositoryName)
-	} else if a.ko.Spec.RepositoryName != nil && b.ko.Spec.RepositoryName != nil {
-		if *a.ko.Spec.RepositoryName != *b.ko.Spec.RepositoryName {
-			delta.Add("Spec.RepositoryName", a.ko.Spec.RepositoryName, b.ko.Spec.RepositoryName)
-		}
-	}
 	if !reflect.DeepEqual(a.ko.Spec.Tags, b.ko.Spec.Tags) {
 		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
 	}
diff --git a/pkg/resource/repository/resource.go b/pkg/resource/repository/resource.go
index e15d755..a2dd27e 100644
--- a/pkg/resource/repository/resource.go
+++ b/pkg/resource/repository/resource.go
@@ -88,7 +88,7 @@ func (r *resource) SetIdentifiers(identifier *ackv1alpha1.AWSIdentifiers) error
 	if identifier.NameOrID == "" {
 		return ackerrors.MissingNameIdentifier
 	}
-	r.ko.Spec.RepositoryName = &identifier.NameOrID
+	r.ko.Spec.Name = &identifier.NameOrID
 
 	f2, f2ok := identifier.AdditionalKeys["registryID"]
 	if f2ok {
diff --git a/pkg/resource/repository/sdk.go b/pkg/resource/repository/sdk.go
index 4366244..61a0053 100644
--- a/pkg/resource/repository/sdk.go
+++ b/pkg/resource/repository/sdk.go
@@ -132,9 +132,9 @@ func (rm *resourceManager) sdkFind(
 			ko.Status.ACKResourceMetadata.ARN = &tmpARN
 		}
 		if elem.RepositoryName != nil {
-			ko.Spec.RepositoryName = elem.RepositoryName
+			ko.Spec.Name = elem.RepositoryName
 		} else {
-			ko.Spec.RepositoryName = nil
+			ko.Spec.Name = nil
 		}
 		if elem.RepositoryUri != nil {
 			ko.Status.RepositoryURI = elem.RepositoryUri
@@ -158,7 +158,7 @@ func (rm *resourceManager) sdkFind(
 func (rm *resourceManager) requiredFieldsMissingFromReadManyInput(
 	r *resource,
 ) bool {
-	return r.ko.Spec.RepositoryName == nil
+	return r.ko.Spec.Name == nil
 
 }
 
@@ -172,9 +172,9 @@ func (rm *resourceManager) newListRequestPayload(
 	if r.ko.Spec.RegistryID != nil {
 		res.SetRegistryId(*r.ko.Spec.RegistryID)
 	}
-	if r.ko.Spec.RepositoryName != nil {
+	if r.ko.Spec.Name != nil {
 		f3 := []*string{}
-		f3 = append(f3, r.ko.Spec.RepositoryName)
+		f3 = append(f3, r.ko.Spec.Name)
 		res.SetRepositoryNames(f3)
 	}
 
@@ -253,9 +253,9 @@ func (rm *resourceManager) sdkCreate(
 		ko.Status.ACKResourceMetadata.ARN = &arn
 	}
 	if resp.Repository.RepositoryName != nil {
-		ko.Spec.RepositoryName = resp.Repository.RepositoryName
+		ko.Spec.Name = resp.Repository.RepositoryName
 	} else {
-		ko.Spec.RepositoryName = nil
+		ko.Spec.Name = nil
 	}
 	if resp.Repository.RepositoryUri != nil {
 		ko.Status.RepositoryURI = resp.Repository.RepositoryUri
@@ -298,8 +298,8 @@ func (rm *resourceManager) newCreateRequestPayload(
 	if r.ko.Spec.RegistryID != nil {
 		res.SetRegistryId(*r.ko.Spec.RegistryID)
 	}
-	if r.ko.Spec.RepositoryName != nil {
-		res.SetRepositoryName(*r.ko.Spec.RepositoryName)
+	if r.ko.Spec.Name != nil {
+		res.SetRepositoryName(*r.ko.Spec.Name)
 	}
 	if r.ko.Spec.Tags != nil {
 		f5 := []*svcsdk.Tag{}
@@ -362,8 +362,8 @@ func (rm *resourceManager) newDeleteRequestPayload(
 	if r.ko.Spec.RegistryID != nil {
 		res.SetRegistryId(*r.ko.Spec.RegistryID)
 	}
-	if r.ko.Spec.RepositoryName != nil {
-		res.SetRepositoryName(*r.ko.Spec.RepositoryName)
+	if r.ko.Spec.Name != nil {
+		res.SetRepositoryName(*r.ko.Spec.Name)
 	}
 
 	return res, nil
```

You will note that there were changes made to the `repository.go` file, the `pkg/resource/repository/sdk.go`.

### Renaming things with different names in input and output shapes

The second reason we might need to rename a field is when the same field goes by different names in the *shapes* (i.e., expected syntax) of the input and output.  An hypothetical example of this might be a field that is called `EnableEncryption` in an input shape and `EncryptionEnabled` in an output shape.  In order to inform the code generator that these fields are actually the same, we would rename one of the fields to match the other.

[**_TODO_** this needs a concrete example of renaming with both `input_fields` and `output_fields`]


## Ignoring things

Sometimes you want to instruct the code generator to simply ignore a particular API Operation, or a particular field in an API Shape.  See [here](https://github.com/aws-controllers-k8s/s3-controller/pull/89) for a real world motivating example of such a need.

You will use the [`ignore:`][ignore-config] block of configuration options to do this.

[ignore-config]: https://github.com/aws-controllers-k8s/code-generator/blob/f6dd767f12429832bc7b4321fb7b763a9fa997c7/pkg/config/config.go#L50-L66

To ignore a specific field in an API Shape, you can list the field via fieldpath in the `ignore.fieldpaths` configuration option.

An [example][ignore-fieldpaths-example] of this can be found in the S3 controller's `generator.yaml` file:

[ignore-fieldpaths-example]: https://github.com/aws-controllers-k8s/s3-controller/blob/9adb7703fa9e8c422a583ec1c8da35ecb21c8917/generator.yaml#L8-L12

```yaml=
ignore:
  field_paths:
    # We cannot support MFA, so if it is set we cannot unset
    - "VersioningConfiguration.MFADelete"
    # This subfield struct has no members...
    - "NotificationConfiguration.EventBridgeConfiguration"
```

When you specify a field path in `ignore.field_paths`, the code generator will skip over that field when inferring custom resource definition `Spec` and `Status` structures.

## Tags

*Most* resources in AWS service APIs can have one or more tags associated with them.  Tags are *typically* simple string key/value pairs; however, the representation of tags across different AWS service APIs is not consistent.  Some APIs use a `map[string]string` to represent tags. Others use a `[]struct{}` where the struct has a `Key` and a `Value` field. Others use more complex structures.

### Telling ACK code generator that a resource does not support tags

There are some API resources that *do not* support tags at all, and we want a way to skip the generation of code that handles tagging for those resources. By default, for all resources, ACK generates some code that handles conversion between the ACK standard representation of tags (i.e., `map[string]string`) and the AWS service-specific representation of tags (e.g., `[]struct{}`, etc).

If you attempt to generate a resource manager for a resource that does not support tags, you will receive an error from the code generator.  ECR's [`PassThroughCacheRule`](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_CreatePullThroughCacheRule.html) is an example of a resource that does not support tags. If we unignore the `PassThroughCacheRule` resource in the ECR controller's [`generator.yaml`](https://github.com/aws-controllers-k8s/ecr-controller/blob/cefc45b65e1a560b9666a1544f26f955c6f84d36/generator.yaml#L78) file and regenerate the controller, we will stumble upon this error:

```
[jaypipes@thelio code-generator]$ make build-controller SERVICE=ecr
building ack-generate ... ok.
==== building ecr-controller ====
Copying common custom resource definitions into ecr
Building Kubernetes API objects for ecr
Generating deepcopy code for ecr
Generating custom resource definitions for ecr
Building service controller for ecr
Error: template: /home/jaypipes/go/src/github.com/aws-controllers-k8s/code-generator/templates/pkg/resource/manager.go.tpl:282:20: executing "/home/jaypipes/go/src/github.com/aws-controllers-k8s/code-generator/templates/pkg/resource/manager.go.tpl" at <.CRD.GetTagField>: error calling GetTagField: tag field path Tags does not exist inside PullThroughCacheRule crd
make: *** [Makefile:41: build-controller] Error 1
```

To fix this error, we used the `tags.ignore` configuration option in [`generator.yaml`](https://github.com/aws-controllers-k8s/ecr-controller/blob/cefc45b65e1a560b9666a1544f26f955c6f84d36/generator.yaml#L78):

```yaml=
ignore:
  resources:
    #- Repository
    #- PullThroughCacheRule
resources:
  Repository:
    renames:
      operations:
        CreateRepository:
          input_fields:
            RepositoryName: Name
        DeleteRepository:
          input_fields:
            RepositoryName: Name
        DescribeRepositories:
          input_fields:
            RepositoryName: Name
  PullThroughCacheRule:
    fields:
      ECRRepositoryPrefix:
        is_primary_key: true
 >  tags:
 >    ignore: true
```

## Resource configuration

### Understanding resource identifying fields

All resources in the AWS world have one or more fields that serve as primary key identifiers. Most people are familiar with the `ARN` fields that most modern AWS resources have. However, the `ARN` field is not the only field that can serve as a primary key for a resource. ACK's code generator reads an API model file and attempts to determine which fields on a resource can be used to uniquely identify that resource. Sometimes, though, the code generator needs to be instructed which field or fields comprise this primary key.  [See below](#Field-level-configuration-of-identifying-fields) for an example from ECR's `PullThroughCacheRule`.

There are resource-level and field-level configuration options that inform the code generator about identifying fields.

#### Resource-level configuration of identifying fields

The `resource[$resource].is_arn_primary_key` configuration option is a boolean, defaulting to `false` that instructs the code generator to use the `ARN` field when calling the "ReadOne" (i.e., "Describe" or "Get") operation for that resource. When `false`, the code generator will look for "identifier fields" with field names such as `ID` or `Name` (along with variants that include the resource name as a prefix, e.g., "BucketName").

Use the `is_arn_primary_key=true` configuration option *when the resource has no other identifying fields*. An example of this is SageMaker's `ModelPackage` resource that has no `Name` or `ID` field and can only be identified via an `ARN` field:

```yaml=
resources:
  ModelPackage:
    is_arn_primary_key: true
```

*[NOTE(jaypipes): Probably want to reevaluate this particular config option and use the field-centric is_primary_key option instead...]*

#### Field-level configuration of identifying fields

Sometimes a resource's primary key field is non-obvious (like `Name` or `ID`). Use the `resources[$resource]fields[$field].is_primary_key` configuration option to tell the code generator about these fields.

An example here is ECR's [`PullThroughCacheRule`](https://github.com/aws-controllers-k8s/ecr-controller/blob/cefc45b65e1a560b9666a1544f26f955c6f84d36/generator.yaml#L57) resource, which [has a primary key field](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_CreatePullThroughCacheRule.html) called `ECRRepositoryPrefix`:

```yaml=
resources:
  PullThroughCacheRule:
    fields:
      ECRRepositoryPrefix:
        is_primary_key: true
```

*[NOTE(jljaco):  If we discard `is_arn_primary_key` in favor of only `is_primary_key`, this sub-section should be moved into the `Field Configuration` section]*

### Correcting exception codes

An ACK controller needs to understand which HTTP exception code means "this resource was not found"; otherwise, the controller's logic that determines whether to create or update a resource falls apart.
 
For the majority of AWS service APIs, the ACK code generator can figure out which HTTP exception codes map to which HTTP fault behaviours. However, some AWS service API model definitions do not include exception metadata. Other service API models include straight-up incorrect information that does not match what the actual AWS service returns.

To address these issues, you can use the `resource[$resource].exceptions` configuration block.

An example of an API model that does not indicate the exception code representing a resource not found is DynamoDB. When calling DynamoDB's [`DescribeTable`](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DescribeTable.html) API call with a table name that does not exist, you will get back a `400` error code instead of `404` and the exception code string is `ResourceNotFoundException`.

To tell the ACK code generator how to deal with this, use the `exceptions` [configuration option](https://github.com/aws-controllers-k8s/code-generator/blob/160b839fe09dd7e1f321e094604ffc3b6ae2a285/pkg/config/resource.go#L251):

```yaml=
resources:
  Table:
    exceptions:
      errors:
        404:
          code: ResourceNotFoundException
```

This configuration instructs the code generator to [produce code](https://github.com/aws-controllers-k8s/dynamodb-controller/blob/eb1405d3d10050c8a866dc0e2dc0ec72c8213886/pkg/resource/table/sdk.go#L79-L84) that looks for `ResourceNotFoundException` in the error response of the API call and interprets it properly as a `404` or "resource not found" error.

#### Specifying terminal codes to indicate terminal state

An `ACK.Terminal` `Condition` is placed on a custom resource (inside of its `Status`) when the controller realizes that, without the user changing the resource's `Spec`, the resource will not be able to be reconciled (i.e., the desired state will never match the actual state).

When an ACK controller gets a response back from an AWS service containing an error code, the controller evaluates whether that error code should result in the `ACK.Terminal` `Condition` being placed on the resource. Examples of these "terminal codes" are things such as: 

- improper input being supplied
- a duplicate resource already existing
- conflicting input values

AWS service API responses having a `4XX` HTTP status code will have a corresponding exception string code (e.g., `InvalidParameterValue` or `EntityExistsException`). Use the `resource[$resource].exceptions.terminal_codes` configuration option to tell the code generation which of these exception string codes it should consider to be a *terminal state* for the resource.

Here is [an example from the RDS controller](https://github.com/aws-controllers-k8s/rds-controller/blob/f97b026cdd72e222390f42a18770fb0de49c3b41/generator.yaml#L97), where we indicate the set of exception string code that will set the resource into a *terminal state*:

```yaml=
resources:
  DBCluster:
    exceptions:
      terminal_codes:
        - DBClusterQuotaExceededFault
        - DBSubnetGroupDoesNotCoverEnoughAZs
        - InsufficientStorageClusterCapacity
        - InvalidParameter
        - InvalidParameterValue
        - InvalidParameterCombination
        - InvalidSubnet
        - StorageQuotaExceeded
```

### Controlling reconciliation and requeue logic

By default, an ACK controller will requeue a resource for future reconciliation only when the resource is in some transitional state.

For example, when you create an RDS [`DBInstance`](https://github.com/aws-controllers-k8s/rds-controller/blob/f97b026cdd72e222390f42a18770fb0de49c3b41/generator.yaml#L184) resource, the resource initially goes into a `CREATING` transitional state and then eventually will arrive at an `AVAILABLE` state. When the RDS controller for ACK initially creates the RDS `DBInstance` resource, it calls the RDS [`CreateDBInstance`](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html) API call, sees the state of the DB instance is `CREATING`, adds an `ACK.ResourceSynced=False` `Condition` to the resource and *requeues* the resource to be processed again in a few seconds.

When the resource is processed in the next reconciliation loop, the controller calls the [`DescribeDBInstance`](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DeleteDBInstance.html) API endpoint and checks to see if the DB instance is in the `AVAILABLE` state. If it is not, then the controller requeues the resource again. If it is in the `AVAILABLE` state, then the controller sets the `ACK.ResourceSynced` `Condition` to `True`, which is the indication to the ACK runtime that the resource should *not* be requeued.

Sometimes, you may want to have the ACK controller requeue certain resources *even after a successful reconciliation loop that leaves the resource in the `ACK.ResourceSynced=True` state*. If this is the case, you should use the `resources[$resource].reconcile.requeue_on_success_seconds` configuration option. The value of this option should be the amount of time (in seconds) after which the reconciler should requeue the resource.

Here is an example of this configuration option as used in the SageMaker controller's [`NotebookInstance` resource](https://github.com/aws-controllers-k8s/sagemaker-controller/blob/c77c6def970cf80322bacaa6aa5ff58dde671dbf/generator.yaml#L595):

```yaml=
resources:
  NotebookInstance:
    # Resource state/status can be modified in Sagemaker Console
    # Need to reconcile to catch these state/status changes
    reconcile: 
      requeue_on_success_seconds: 60
```

We set this `requeue_on_success_seconds` value to `60` here because the values of various fields in this Sagemaker resource tend to change often and we want the `Status` section of our custom resource to contain values that are fresher than the default requeue period (10 hours as of this writing).

## Field configuration

**TODO**

### Marking a field as required

**TODO**

### Controlling a field's Go type

**TODO**

### Controlling how a field's values are compared

**TODO**

### Mutable vs. immutable fields

**TODO**

### Controlling where a field's definition comes from

**TODO**

### Annotating a field as a "Printer Column"

**TODO**

### Controlling field "late initialization"

**TODO**

### Informing the code generator that a field refers to another Resource

**TODO**

### Adding custom fields

**TODO**

## Custom code hook points

The code generator will generate Go code that implements the `aws-sdk-go` SDK "binding" calls. Sometimes you will want to inject bits of custom code at various points in the code generation pipeline.

Custom code [hook points](https://github.com/aws-controllers-k8s/code-generator/blob/160b839fe09dd7e1f321e094604ffc3b6ae2a285/pkg/generate/ack/hook.go#L28) do this injection. They should be preferred versus using complete overrides (e.g.,  `resources[$resource].update_operation.custom_method_name`). The reason that custom code hooks are preferred is because generally you want to maximize the amount of *generated* code and minimize the amount of *hand-written* code in each controller.  *[NOTE(jljaco): decide later whether to bother documenting complete overrides via `update_operation.custom_method_name`]*

**TODO**

## Attribute-based APIs

**OMG TODO.**

## Miscellaneous/maybe cover later/documentation backlog

### What does PrefixConfig do?
### What if the code generator cannot figure out my service's API model name?
### list_operation.match_fields
