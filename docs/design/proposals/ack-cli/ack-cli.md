# ack-cli: command surface

> **Status: proposal.** Opened for community feedback; nothing here has shipped.

`ack-cli` is the operator-facing ACK command line tool — for people running ACK
against their own AWS account. Contributor tooling (`ackdev`) and code generation
(`ack-generate`) serve ACK contributors and remain separate binaries.

This doc is the command surface only. `adopt` is specified in
[adopt-by-tags.md](adopt-by-tags.md).

---

## Commands

| Command | Purpose | AWS calls |
|---------|---------|-----------|
| `adopt` | Bring existing AWS resources under ACK management, discovered by tag | `tag:GetResources` |
| `list adoptable` | Which ACK kinds can be adopted by tag, what identifiers each needs, and why the rest cannot | none |
| `version` | CLI version and the resource catalog it was built from | none |

Nothing writes to a cluster. `adopt` emits YAML on stdout for review, and
`list adoptable` reads a catalog embedded in the binary, so it works offline.

### `adopt`

The reason the tool exists. ACK adopts an existing resource when you give it a CR
carrying `services.k8s.aws/adoption-fields` — the identifier values the controller
needs to find that resource. Authoring that by hand means knowing each kind's
identifier keys and extracting the values yourself, for every resource. `adopt`
derives them from a tag selector.

```console
$ ack-cli adopt --service eks --kind Nodegroup --adoption-set platform-prod \
    --tags "Environment=prod" | kubectl create -f -
```

It targets one kind per run and applies nothing itself. See
[adopt-by-tags.md](adopt-by-tags.md) for flags, emitted output, and rollout.

### `list adoptable`

What can be adopted and what each kind needs — the reference a user reads before
running `adopt`, and the honest account of what the tool cannot do. **207 of ACK's
261 resources** are adoptable by tag.

```console
$ ack-cli list adoptable --service eks
SERVICE  KIND                    TYPE FILTER                 IDENTIFIER KEYS
eks      Addon                   eks:addon                   name,clusterName
eks      Capability              eks:capability              name,clusterName
eks      Cluster                 eks:cluster                 name
eks      FargateProfile          eks:fargateprofile          clusterName,name
eks      IdentityProviderConfig  eks:identityproviderconfig  clusterName
eks      Nodegroup               eks:nodegroup               clusterName,name
```

`TYPE FILTER` is the string sent to the Resource Groups Tagging API, which makes a
runtime "type not supported" error legible. `IDENTIFIER KEYS` are the fields
`adopt` will populate — what the user would otherwise hand-author.

`--unsupported` lists the 54 kinds that cannot be adopted by tag, each with a
reason, so a missing kind is explained rather than silently absent:

```console
$ ack-cli list adoptable --unsupported --service eks
RESOURCE                      KIND                    REASON
eks:access_entry              AccessEntry             identifier not encoded in the ARN
eks:pod_identity_association  PodIdentityAssociation  identifier not encoded in the ARN
```

### `version`

Reports the CLI version and its catalog, which is compiled into the binary — so
coverage is a property of the build, not of ACK today. The version *is* the
`ack-chart` release the catalog was generated from.

```console
$ ack-cli version
ack-cli 1.0.42  (ack-chart 1.0.42)
resource catalog: 207 adoptable of 261 known resources
```
