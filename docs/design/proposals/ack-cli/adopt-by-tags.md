# `ack-cli adopt`: adoption by tags

`ack-cli` is a local command-line tool for ACK. It queries AWS APIs and writes
Kubernetes manifests to stdout, so it needs no cluster access; `adopt` and
`list adoptable` are its commands today.

This proposal covers `adopt`: how bulk adoption works, what each flag means, and how
the CLI is released as new ACK resources land upstream.

---

## 1. What it does

ACK adopts an existing AWS resource when you hand it a CR carrying
`services.k8s.aws/adoption-fields` — a JSON map of the identifier values the
controller needs to find that resource. Authoring it by hand means knowing each
kind's identifier keys and pulling the values out of every resource yourself.

`adopt` derives that map. Given a tag selector it discovers matching resources,
reads each identifier value out of the resource's ARN, and emits adoption
manifests:

```
--tags "Environment=prod" --service eks --kind Nodegroup
        │
        ▼  GetResources (Resource Groups Tagging API), scoped to eks:nodegroup
  arn:aws:eks:us-west-2:111122223333:nodegroup/prod-cluster/ng-general/a1b2-uuid
        │  match against the kind's ARN template
        │    nodegroup/${ClusterName}/${NodegroupName}/${UUID}
        ▼  read each identifier out of its bound placeholder
  {"clusterName": "prod-cluster", "name": "ng-general"}
        │
        ▼  emit CR: adoption-fields + policy annotations + region
```

**One kind per run**, named with `--service` and `--kind`. A broad tag selector
would otherwise sweep in kinds the user never meant to hand to ACK, and one
`kubectl create` stream would mix kinds whose adoption semantics differ. Naming the
kind also removes any need to infer it from an ARN — flat ARNs like `sns:topic`
carry no type label and shared prefixes like `rds:db:…` are ambiguous across kinds.

**Nothing is applied.** Manifests go to stdout; the summary, per-resource skips and
`--debug` go to stderr. The set can be reviewed before anything reaches the cluster,
then piped through once it looks right:

```console
$ ack-cli adopt --service eks --kind Nodegroup --tags "Environment=prod" \
    | kubectl create -f -
```

Writing to the cluster is left to `kubectl`, which already handles context selection,
namespace defaults and RBAC errors.

**Use `create` (POST), not `apply` (patch).** `create` is atomic, and the API server
rejects a duplicate name with `AlreadyExists`, so one AWS resource cannot end up
adopted by two CRs. Rejection is per object rather than per batch: a run across a
partly-adopted set creates the new CRs, reports `AlreadyExists` for the rest, and
adopts only the delta.

`apply` would undo both properties. It patches existing CRs to match the latest run,
which re-asserts `read-only: "true"` over a set already handed to ACK (§2.3). It also
treats the manifests as desired state, when they are only a starting point: the
controller populates `spec` from the live resource once adoption succeeds.

**Scope: 207 of ACK's 261 resources** are adoptable this way. The other 54 are not
discoverable by tag or not decomposable from their ARN; `ack-cli list adoptable
--unsupported` reports each with a reason, and they remain adoptable one at a time
by hand-writing the `adoption-fields` annotation per the [ACK adoption
docs][adoption].

[adoption]: https://aws-controllers-k8s.github.io/docs/guides/adoption

---

## 2. Flags

```
ack-cli adopt --service SERVICE --kind KIND --tags "K=V ..." [flags]
```

| Flag | Meaning |
|------|---------|
| `--service` | ACK service, e.g. `eks`. **Required** |
| `--kind` | ACK resource kind, e.g. `Nodegroup`. **Required** |
| `--tags "K=V K=V"` | Tag selector, given once as space-separated pairs. All must match (AND). A bare `KEY` matches any value. **Required** |
| `--adoption-set NAME` | Names the collection this adopts; applied as the `ack.k8s.aws/adoption-set` label. Generated from the service, kind and a short run suffix when omitted. Does **not** affect CR names |
| `--policy` | ACK adoption policy: `adopt` (default). `adopt-or-create` is refused (§2.3) |
| `--read-only` | Emit `services.k8s.aws/read-only`. Default `true`; pass `--read-only=false` to let ACK manage the resource (§2.3) |
| `--deletion-policy` | Emit `services.k8s.aws/deletion-policy`: `retain` (default) or `delete` (§2.3) |
| `--region` | AWS region. Defaults to the environment; fails if none resolves (§2.2) |
| `--namespace` | Namespace on emitted CRs |
| `--debug` | Log discovery and resolution to stderr (also `ACK_CLI_DEBUG=1`) |

Output always goes to stdout; writing to a file is a shell redirect
(`> nodegroups.yaml`).

IAM: `tag:GetResources`, plus `sts:GetCallerIdentity` under `--debug` to report
which identity is querying. `adopt` never contacts a cluster, so it needs no
kubeconfig and no Kubernetes permissions.

### 2.1 `--tags`

Given **once**, with space-separated `key=value` pairs. It is a filter — every pair
must match, so each one narrows the result:

```console
# nodegroups tagged Environment=prod AND team=platform
$ ack-cli adopt --service eks --kind Nodegroup --adoption-set platform-prod \
    --tags "Environment=prod team=platform"

# bare key: any nodegroup carrying an `Environment` tag, whatever its value
$ ack-cli adopt --service eks --kind Nodegroup --adoption-set platform-prod \
    --tags "Environment"
```

Tag keys and values are case-sensitive. An empty result has three causes a user
cannot tell apart — wrong region or account, no resources of the kind at all, or
resources that exist without the tags — so `adopt` re-queries without the tag filter
and reports which case it hit, including the tag keys that *are* present on
resources of that kind.

### 2.2 `--region`

Region resolves in this order, and **fails if nothing resolves** rather than
defaulting to a region:

```
--region  →  AWS_REGION  →  AWS_DEFAULT_REGION  →  region in the active AWS profile
```

The resolved region is emitted as `services.k8s.aws/region` on every CR. This is a
correctness requirement, not provenance: the ACK runtime reads that annotation in
preference to the namespace's `default-region` and the controller's own region, and
explicitly will not override it. Without it, a CR generated from a `us-west-2`
query and applied to a cluster whose controller defaults to `us-east-1` sends the
controller looking for `prod-cluster/ng-general` in the wrong region — failing, or
adopting a same-named resource that happens to exist there.

**The annotation is written from the query region, not the ARN.** `GetResources` is
regional, so the queried region is authoritative for everything it returns —
including resources whose ARN omits the region. 22 of the 207 adoptable ARN
templates have an empty region slot (`iam:role`, `route53:hosted_zone`,
`cloudfront:distribution`, `s3:bucket`, `rds:global-cluster`, …), and `s3:Bucket` is
regional despite that, so the region is never inferred per resource.

`services.k8s.aws/owner-account-id` is deliberately **not** emitted. The same
argument would apply, but that annotation drives ACK's CARM path and on a cluster
with no CARM role configuration it can stop the controller resolving a role.

### 2.3 `--policy`, `--read-only`, `--deletion-policy`

ACK uses a separate annotation for each of three things: how the resource is found,
whether ACK may change it, and what happens when its CR is deleted. Each one gets its
own flag.

| Flag | Default | Emitted | Effect |
|------|---------|---------|--------|
| `--policy` | `adopt` | `services.k8s.aws/adoption-policy: adopt` | The controller finds the resource by its adoption-fields and brings it under management |
| `--read-only` | `true` | `services.k8s.aws/read-only: "true"` | ACK reads the resource and reports status, but never mutates it |
| `--deletion-policy` | `retain` | `services.k8s.aws/deletion-policy: retain` | `kubectl delete` on the CR removes only the CR; the AWS resource survives |

**The defaults make the first run observe-only.** Adopting infrastructure ACK did not
create should not, on its own, give ACK permission to change or destroy it, so both are
opt-in. A too-broad selector or a mis-mapped identifier then costs a CR to delete
instead of a production resource. Handing over management is a second, deliberate step,
and the adoption-set label makes it one command:

```console
$ kubectl annotate nodegroup -l ack.k8s.aws/adoption-set=platform-prod \
    services.k8s.aws/read-only=false --overwrite
```

Both can also be set when the manifests are generated: `--read-only=false
--deletion-policy=delete` emits a fully managed CR.

ACK's other adoption policy, `adopt-or-create`, is refused. It creates the resource
from `spec` when none is found, and these manifests carry no spec (§3), so the error
says that rather than reporting a generic invalid value.

---

## 3. What gets emitted, and what it guarantees

```yaml
apiVersion: eks.services.k8s.aws/v1alpha1
kind: Nodegroup
metadata:
  name: prod-cluster-ng-general-7f3a9c1d
  namespace: default
  labels:
    ack.k8s.aws/adoption-set: platform-prod    # added by the CLI; selects this set
  annotations:
    services.k8s.aws/adoption-fields: '{"clusterName":"prod-cluster","name":"ng-general"}'
    services.k8s.aws/adoption-policy: adopt
    services.k8s.aws/read-only: "true"
    services.k8s.aws/deletion-policy: retain
    services.k8s.aws/region: us-west-2
```

- **`spec` is omitted entirely**, rather than emitted as `spec: {}`. The controller
  fills it in from the live resource while adopting. ACK's generated types declare
  `spec,omitempty`, so a CR with no spec validates; `spec: {}` is present-but-empty,
  gets checked against spec's own required fields, which most kinds have, and is
  rejected by the API server.
- **`--adoption-set` is the handle for the collection**, generated when the flag is
  omitted. It is a label, never part of a name, so one value can span several runs and
  still select them all — which is also how a user verifies the adoption took:

  ```console
  $ kubectl wait --for=condition=ACK.ResourceSynced nodegroup \
      -l ack.k8s.aws/adoption-set=platform-prod --timeout=5m
  ```

- **Names derive from resource identity** — the identifier values plus a digest of
  the source ARN — so re-adopting regenerates the same name and `kubectl create`
  rejects it. ACK has no cross-CR ownership guard: two CRs naming one AWS resource
  would both reconcile it, and once either leaves read-only they would compete over
  it. The digest separates same-named resources in different regions or accounts.
- **Unresolvable ARNs are skipped, not guessed at.** An ARN that does not match its
  kind's template is skipped with a reason on stderr and counted in the summary. A
  partial identifier map is never emitted, since it would adopt the wrong resource.
- **Output is deterministic.** `GetResources` is paginated and promises no
  particular order, so CRs are sorted by name before rendering and `adoption-fields`
  is key-sorted. The same inputs produce byte-identical output, which keeps the
  manifests reviewable as a diff when they are checked into Git.

---

## 4. Where the mapping comes from

The ARN-to-identifier mapping is generated offline into a catalog that is committed
and embedded in the binary (`go:embed`), so it is reviewable as a diff and the CLI's
only network dependency is AWS itself.

```jsonc
{
  "resources": [
    {
      "service": "eks", "resource_dir": "nodegroup", "kind": "Nodegroup",
      "group": "eks.services.k8s.aws", "version": "v1alpha1",
      "resource_type_filter": "eks:nodegroup",
      "arn_template": "arn:${Partition}:eks:${Region}:${Account}:nodegroup/${ClusterName}/${NodegroupName}/${UUID}",
      // ORDER = ACK's declaration order. `key` is the adoption-fields key,
      // `from` is where its value is read out of the ARN.
      "bindings": [
        {"key": "clusterName", "from": "${ClusterName}"},
        {"key": "name",        "from": "${NodegroupName}"}
      ]
    }
  ],
  "unsupported": [
    {"resource": "eks:access_entry", "kind": "AccessEntry",
     "reason": "identifier not encoded in the ARN"}
  ]
}
```

`bindings` is the adoption contract for a kind: the keys `adoption-fields` must
carry, in order, and where each value comes from. `from` is a template over the ARN
template's placeholders — usually a single one, as above, but it can combine several
(§4.1). The same form covers two cases that would otherwise be special: a key holding
the AWS account is `"${Account}"`, and a kind whose identifier *is* its ARN is one
binding `{"key": "arn", "from": "${ARN}"}`, where `${ARN}` stands for the whole
matched ARN.

Two inputs, both machine-read, neither requiring anything of ACK contributors:

1. **Identifier keys and GVK**, parsed from each controller's generated
   `pkg/resource/<r>/resource.go` (`PopulateResourceFromAnnotation` *is* the
   definition of what `adoption-fields` must contain) and `descriptor.go`, read at
   each repo's highest stable tag ≥ v0.1.0 so the catalog only describes released
   controllers.
2. **ARN grammar**, from the AWS Service Reference API — per-resource ARN templates
   with named placeholders.

Keys bind to placeholders **by name**, after normalizing away the repetition of the
resource type that the two sides spell differently — an `ecr:repository`'s
`${RepositoryName}` placeholder binds to its `name` key. Matching by name rather than
position keeps the mapping independent of segment order: a wafv2 web ACL's ARN is
`${Scope}/webacl/${Name}/${Id}` while its ACK keys are `[name, id, scope]`.

A resource is adoptable only if **every** identifier key binds and it has a type
filter; anything else goes to `unsupported` with a reason. A partial mapping is
never shipped, because without a type filter the query cannot be scoped and without
every key the resulting `adoption-fields` would point at the wrong resource.

That yields **207 adoptable of 261 known resources**, across 68 tagged controller
repos.

### 4.1 Overrides

An override is a hand-written catalog entry for a kind that automatic name matching
cannot map. It supplies the type filter, the ARN template, and the bindings, in the
same shape the generator emits. The table holds 24 entries today and is maintained
by this project — **contributors adding a new ACK resource never touch it.** The
generator validates that every override binds all of its kind's identifier keys and
fails the build if one does not, so a typo cannot ship a wrong mapping.

Three cases justify one:

**Cryptic AWS type labels.** The RDS family uses abbreviations no normalization
recovers — `pg`, `subgrp`, `cluster-pg` for parameter groups and subnet groups.

**ACK service aliases.** The ACK service name is not the AWS namespace:
`documentdb` resources are `rds:*` ARNs, `eventbridge:Rule` is `events:rule`,
`prometheusservice` is `aps`. The type filter must use the AWS namespace or the
Tagging API never matches.

**Composed identifiers.** The identifier is in the ARN but spread across segments.
`eks:AccessEntry` is the example. Its ARN carries the principal's type, account and
name as three separate segments:

```
arn:aws:eks:us-west-2:111122223333:access-entry/my-cluster/role/111122223333/my-role/<uuid>
```

| ARN placeholder | value here | feeds |
|-----------------|------------|-------|
| `${ClusterName}` | `my-cluster` | `clusterName` |
| `${IamIdentityType}` | `role` | `principalARN` |
| `${IamIdentityAccountID}` | `111122223333` | `principalARN` |
| `${IamIdentityName}` | `my-role` | `principalARN` |

ACK wants the principal as one ARN, so three of those segments have to be
reassembled into `"principalARN": "arn:aws:iam::111122223333:role/my-role"`:

```jsonc
"eks:access_entry": {
  "resource_type_filter": "eks:access-entry",
  "arn_template": "arn:${Partition}:eks:${Region}:${Account}:access-entry/${ClusterName}/${IamIdentityType}/${IamIdentityAccountID}/${IamIdentityName}/${UUID}",
  "bindings": [
    {"key": "clusterName",  "from": "${ClusterName}"},
    {"key": "principalARN", "from": "arn:${Partition}:iam::${IamIdentityAccountID}:${IamIdentityType}/${IamIdentityName}"}
  ]
}
```

Composed bindings need a real-ARN test on top of the round-trip check (§5), since a
composition can match the template and still be wrong: an IAM role with a path, or an
SSO role, does not render the same way.

**When not to write one.** If the identifier is not in the ARN in any form, no
template can recover it and an override would be a guess that silently adopts the
wrong resource. Those kinds stay unsupported with that reason —
`eks:PodIdentityAssociation` (`associationID`), `sns:Subscription`, and
sub-resources that carry no tags of their own (`route53:RecordSet`, `kms:Grant`,
`efs:MountTarget`) — and are adopted individually by hand instead.

---

## 5. Rollout: releasing the CLI as resources land upstream

New controllers release continuously, and a resource missing from the embedded
catalog simply does not appear in `list adoptable`. The catalog is generated wholly
from public inputs, so refreshing it is mechanical and can be automated end to end.

**Trigger: a new tag in `ack-chart`.** `aws-controllers-k8s/ack-chart` aggregates
every released controller and is re-tagged by `ci-robot` on each controller release,
which makes it the single signal that upstream resources changed. A prow postsubmit
in `test-infra` fires on its semver tags (`^[0-9]+\.[0-9]+\.[0-9]+$`), the same
trigger the existing `ack-chart-release` job uses.

**The job runs a release script:**

1. Regenerate the catalog — enumerate `*-controller` repos, take each one's highest
   stable tag ≥ v0.1.0, parse `resource.go` and `descriptor.go`, fetch ARN grammar
   from the Service Reference API, apply overrides.
2. Validate. The round-trip validator fills every placeholder in each ARN template
   with a unique sentinel value, re-parses the resulting synthetic ARN, and asserts
   that each key recovered *its own* sentinel. That catches the mistakes a name-match
   check cannot see, because a wrong binding still looks like a successful match: one
   key reading another's segment, a segment miscounted against the ARN's
   partition/region/account header, or a mapping that only works when the keys happen
   to be declared in ARN order. An unbound key, an `arn_template` with no type
   filter, or an entry missing its GVK or bindings fails the job.
3. Run `go test ./...`.
4. Build and publish the CLI, tagged with the `ack-chart` version that triggered the
   run.

**The `ack-chart` version is the CLI version.** `ack-cli 1.0.42` was built from
`ack-chart 1.0.42`, so "which controllers does my binary know about?" has a single
answer, and `ack-cli version` reports it. No separate versioning scheme to
reconcile.

Every step must pass before publish, so a bad mapping fails the release instead of
shipping. The generator is deterministic, so a chart tag that changes no resource
produces an identical catalog and a release differing only in version. The regenerated
catalog lands in the release commit, so any change to a binding is visible in that
diff.

A new upstream resource therefore needs no maintainer work: the controller releases,
`ack-chart` is re-tagged, the job regenerates and validates, and the resource shows
up in `list adoptable` in the next CLI release. Work is only needed when a resource
lands in `unsupported` for a reason an override could fix (§4.1).
