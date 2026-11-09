# Hybrid Monorepo with Publishing

This proposal seeks to consolidate ACK development back into a single source
repository while preserving the per-service published repositories that exist
today. Development happens in one place; published artifacts (the
`$SERVICE-controller` repos) continue to exist as the user-facing surface,
populated by an automated publishing bot.

This is a partial reversal of [split-repo.md](../split-repo/split-repo.md) — it
keeps the parts that worked (per-service tagging, per-service release cadence,
stable import paths) and undoes the parts that have become painful (development
spread across 30+ repos).

## Background

In [split-repo.md](../split-repo/split-repo.md), ACK moved from a single source
repository to a constellation of repositories: `code-generator`, `runtime`,
`community`, and one `$SERVICE-controller` repository per AWS service. That
split solved two real problems:

1. **Release series inflexibility** — every service needed independent
   SemVer tags and independent release cadence. Git tags are repo-scoped, so
   one repo per service was the obvious answer.
2. **Common runtime import inflexibility** — service controllers needed to
   pin a specific version of `runtime` via `go.mod`, which required `runtime`
   to be its own taggable repo.

Both problems are still real. Any successor design has to preserve those
properties.

## The Problem

The split-repo model solved release flexibility but introduced friction in
day-to-day development that has compounded as the project has grown to 70+
service controllers.

### Problem #1: Cross-cutting changes are slow and lossy

A change to `runtime` or `code-generator` that affects all controllers
requires:

1. Merge change in `runtime` (or `code-generator`), cut a release.
2. Open ~70 individual PRs against `$SERVICE-controller` repos to bump the
   dependency and regenerate code.
3. Wait for ~70 CI runs, ~70 reviews, ~70 merges, ~70 releases.

This makes refactors that should take an afternoon take weeks. It also makes
it easy for individual services to drift: a service that doesn't get its bump
PR merged in a timely fashion falls behind, and over time the fleet ends up on
N different runtime versions with subtly different behavior.

### Problem #2: No atomic refactors across services

A change that needs to touch the generator, the runtime, and a couple of
service controllers together cannot be made atomically. The PR has to be
broken into a chain of dependent PRs across repos, each waiting on the
previous to release before the next can land. Reviewers cannot see the full
change in one place.

### Problem #3: CI and tooling duplication

Each `$SERVICE-controller` repo carries near-identical CI workflows, OWNERS
plumbing, Makefiles, release scripts, and Helm chart scaffolding. Fixing or
improving any of these is a fan-out problem. In practice the `test-infra`
repo already centralizes the Prow job definitions, but the per-repo files
(`.github/workflows/*`, `Makefile`, `helm/`) still drift.

### Problem #4: Contributor friction

New contributors don't know which repo to file an issue against, which repo
to PR against, or how to test a change that crosses service boundaries.
Cloning a meaningful subset of ACK requires cloning the long tail of repos.

### Problem #5: Development discoverability

There is no single git history that captures the evolution of ACK as a
system. Cross-service patterns, regressions, and bisects are harder than they
need to be because the history is fragmented.

## Prior art: Kubernetes `staging/` + publishing-bot

The upstream Kubernetes project solves a similar shape of problem. The main
`kubernetes/kubernetes` repository contains a `staging/src/k8s.io/` directory
with subtrees for `client-go`, `api`, `apimachinery`, and roughly thirty
other libraries. Development happens in `kubernetes/kubernetes`. Vendoring
back into k/k itself is done with symlinks under `vendor/`.

The standalone repos that users actually consume — `github.com/kubernetes/client-go`,
`github.com/kubernetes/api`, and so on — are populated by a tool called
[`publishing-bot`](https://github.com/kubernetes/publishing-bot). The bot
runs on a schedule, reads a rules file
(`staging/publishing/rules.yaml`) that declares for each target repo which
staging directory to source, which branches to publish, which Go version to
pin, and which dependencies to rewrite. For each rule it filter-rewrites
commits from k/k, regenerates `go.mod`, runs smoke tests, and force-pushes the
synthesized history to the target repo.

The result: developers work in one place, consumers see stable per-library
repos with their own tags, imports, releases, and (read-only) issue trackers.

This is the model this proposal adapts for ACK.

## Scope

In scope for this proposal:

- All `$SERVICE-controller` repositories
- `runtime`
- `code-generator`
- `pkg` (shared utilities — `compare`, `names`, `path`, `strutil`)

Out of scope (for now): `controller-bootstrap`, `ack-chart`, `dev-tools`,
`test-infra`, and other auxiliary repos. They continue to live as
standalone repos and may be revisited later.

## Solution

We propose the following changes:

1. **Rename `github.com/aws-controllers-k8s/community` to
   `github.com/aws-controllers-k8s/ack`.** This repository becomes the
   development monorepo. GitHub redirects preserve existing links, issues,
   stars, and activity. Governance, docs, and design proposals that live in
   `community/` today continue to live in the renamed `ack/` repo.

2. **Pull `pkg`, `runtime`, `code-generator`, and all service
   controllers into the monorepo** as subdirectories, each retaining its
   own `go.mod`. A top-level `go.work` ties them together. Everything
   that is mirror-published lives under a single `staging/` directory:

   ```
   ack/
     go.work                          # workspace tying all modules together
     docs/                            # monorepo-only
     scripts/                         # monorepo-only
     publishing/
       rules.yaml                     # one rule per publish target
     staging/                         # everything here is mirror-published
       pkg/                           # -> aws-controllers-k8s/pkg
       runtime/                       # -> aws-controllers-k8s/runtime
       code-generator/                # -> aws-controllers-k8s/code-generator
       services/
         s3-controller/               # -> aws-controllers-k8s/s3-controller
         iam-controller/              # -> aws-controllers-k8s/iam-controller
         ...
   ```

   Each directory under `staging/` is one Go module with its own `go.mod`
   and maps 1:1 to a downstream publish-target repo — the directory
   basename equals the repo name. Libraries sit directly under
   `staging/`; the 70+ service controllers are grouped under
   `staging/services/`. Everything outside `staging/` (`docs/`,
   `scripts/`, `publishing/`) is monorepo-only and never mirrored.

   Each `staging/services/$SERVICE-controller/` directory contains
   exactly what the corresponding `$SERVICE-controller` repo contains
   today — `apis/`, `pkg/`, `helm/`, `config/`, `test/`, the works.
   Layout inside a controller directory does not change.

   The `staging/` prefix is a single structural invariant: a directory
   is a publish target if and only if it lives under `staging/`. This
   keeps every `rules.yaml` source path on a common prefix and makes
   `ack_validate_rules` trivial. Independent `go.mod` files also preserve
   per-controller Go version pins and dependency upgrade cadence.

3. **The existing `$SERVICE-controller`, `runtime`, and `code-generator`
   repositories continue to exist.** They become **bot-write-only publish
   targets**. Their tags, releases, GitHub releases, image registries, Helm
   chart publication, and Go import paths are unchanged. Users who consume
   `github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1` see no
   difference.

4. **A publishing bot mirrors monorepo subtrees out to the per-service
   repos.** We reuse the upstream `kubernetes/publishing-bot` binary and
   container image directly — no fork. The bot is driven by a `rules.yaml`
   under `publishing/` in the monorepo that declares for each target:
   source directory, target branches, smoke-test command, and any path or
   dependency rewrites. The bot writes a synthesized `go.mod` for each
   published target, rewriting the in-tree workspace replace directives
   into SemVer-tagged dependencies (so `runtime` is consumed in-tree by
   workspace replace during development, but appears as a `go.mod`
   dependency on a tagged `runtime` release in the published service repo).

5. **Release PRs happen in the monorepo, against the relevant
   `staging/services/$SERVICE-controller/helm/values.yaml`.** When the
   release PR merges, the
   publishing bot mirrors the commit (including the `helm/values.yaml`
   bump) to `$SERVICE-controller:main`. The existing
   `{service}-controller-release-tag` postsubmit fires on the mirrored push,
   reads `helm/values.yaml`, validates the bump, and tags
   `$SERVICE-controller` with the new SemVer. From that point the existing
   release machinery — `create-release.yml` GitHub Action,
   `release-controller.sh` postsubmit, `olm-bundle-pr`, `update-ack-chart` —
   fires exactly as it does today.

6. **The `code-generator` build script is adapted to run from a monorepo
   root.** Today `code-generator/scripts/build-controller.sh` hardcodes
   `../$SERVICE-controller`. It will honor `MONO_REPO_ENABLED` and, when
   set, resolve the controller source as
   `staging/services/$SERVICE-controller` within the monorepo. The
   `ack-generate` CLI already accepts all required paths as flags, so no
   Go-level changes are needed.

7. **Per-service repos become bot-write-only.** Branch protection on every
   `$SERVICE-controller` repo is tightened so that only the publishing bot
   can push to `main`. Human PRs opened against `$SERVICE-controller` repos
   are auto-closed by a bot that posts a redirect link to the monorepo. The
   `controller-release-tag` and downstream release jobs continue to run on
   pushes/tags in the per-service repo as today.

8. **Issues stay where they already are.** ACK already centralizes issues
   in the `community` repo today; once renamed to `ack`, this is unchanged.
   No issue migration is needed.

### What this preserves from split-repo

The two problems that motivated split-repo remain solved:

- **Per-service release series** — `$SERVICE-controller` repos still hold
  the release tags. A team can keep `main` and `stable` branches on their
  per-service repo, and the publishing bot can be configured to mirror more
  than one monorepo branch into more than one target branch.
- **Versioned runtime imports** — the `runtime` repository continues to
  exist as a publish target. Service controllers (as observed by consumers
  and the Go module system) still import a SemVer-tagged `runtime`. Inside
  the monorepo, services consume `staging/runtime/` in-tree via the Go
  workspace, and the publishing bot rewrites that to a tagged dependency
  in each published service repo's `go.mod`.

## Release flow under the new model

```
[ developer ]
     |
     |  PR against ack/staging/services/s3-controller/helm/values.yaml (image.tag: v1.2.0)
     v
[ ack monorepo: PR merges to main ]
     |
     |  publishing-bot picks up the commit on its next tick,
     |  filter-rewrites it onto s3-controller:main,
     |  force-pushes (preserving tags)
     v
[ s3-controller:main push ]
     |
     |  Prow postsubmit `s3-controller-release-tag` fires,
     |  reads helm/values.yaml, validates next-patch-or-minor,
     |  pushes tag v1.2.0 to s3-controller
     v
[ s3-controller: v1.2.0 tag ]
     |
     |--> GitHub Action create-release.yml -> GitHub release
     |--> Prow postsubmit s3-post-submit  -> image+chart build & publish
     |--> Prow postsubmit s3-olm-bundle-pr -> OLM bundle PR
     |--> Prow postsubmit update-ack-chart -> umbrella chart update
```

The release-tag script reads `helm/values.yaml` from the per-service repo
checkout and does not know or care that the upstream commit came from a
monorepo mirror. No changes to the release-tag, release-controller, OLM, or
ack-chart pipelines are required.

## Bot trigger model

The publishing bot is a binary; nothing about its design requires a cron
schedule. Upstream `kubernetes/publishing-bot` is deployed as a `CronJob` for
batching reasons, but ACK has tighter release-latency expectations (a release
PR merge should not have to wait an hour for the next tick). We propose a
**hybrid trigger model**:

1. **On-merge postsubmit (primary trigger).** A Prow postsubmit on the
   monorepo's `main` branch invokes the publishing bot on every merge. The
   job inspects the changed paths in the merge commit and runs only the
   rules whose source directory was touched — a merge that only edits
   `staging/services/s3-controller/` produces only an `s3-controller`
   mirror push, not a fleet scan. This keeps median propagation latency
   in the minutes range and
   makes release PRs feel responsive.

2. **Cron heartbeat (backup).** A low-frequency cron run (e.g. hourly)
   executes the bot across all rules. This catches anything that a failed
   postsubmit missed and protects against drift if a postsubmit run was
   skipped or aborted. The cron run is also the path by which
   `runtime`/`code-generator` republishes are picked up if their rules
   include time-based triggers.

3. **Concurrency lock.** The bot acquires a lock (a Kubernetes Lease, or an
   advisory lock on a known object) before pushing. Two near-simultaneous
   monorepo merges should not race each other to force-push the same
   downstream branch. Subsequent invocations wait, observe the new
   upstream `HEAD`, and replay cleanly.

4. **Failure handling.** A failed on-merge run is alerted via the existing
   Prow alert channels. The cron heartbeat provides eventual consistency
   even if the postsubmit alert is missed; explicit re-runs are possible via
   the standard Prow re-run UX.

This model gives upstream-style robustness (the cron will always reconcile)
with the responsiveness needed for the release PR flow.

## Configuration: tracking migrated components

Two distinct pieces of state drive the migration. They answer different
questions and are deliberately kept separate.

### `monorepo_components` — which components have migrated

`test-infra`'s `jobs_config.yaml` already lists every service in a flat
`aws_services` array and maintains several subset lists
(`code_gen_presubmit_services`, `runtime_presubmit_services`, …) that the
job generator tests membership against. The migration adds one more such
list:

```yaml
aws_services:
- acm
- s3
- iam
- ...

# components whose source now lives in the ack monorepo
monorepo_components:
- pkg
- runtime
- code-generator
- s3
- iam
```

`monorepo_components` is a flat membership oracle covering both service
controllers and the library modules (`pkg`, `runtime`,
`code-generator`) — none of which live in `aws_services`. The generator
tests membership exactly as it already does for the soak and codegen
subsets:

- service templates iterate `aws_services` and check
  `contains .MonorepoComponents $service`
- `runtime_tests.tpl` / `pkg_tests.tpl` / `code_generator_tests.tpl`
  check `contains .MonorepoComponents "runtime"` (etc.)

A component in the list gets the monorepo job variant; one absent gets
the standalone variant. This is the per-component migration ledger on the
Prow side, and it is what makes the wave migration incremental — each
wave is one edit to this list. It pairs with `publishing/rules.yaml` in
the monorepo, the equivalent ledger on the publishing-bot side; the
`ack_validate_rules` presubmit keeps the two consistent.

As the migration completes, `monorepo_components` grows until it covers
everything, at which point it and `aws_services` collapse and the
standalone templates are deleted (Stage 8).

### `MONO_REPO_ENABLED` — telling scripts which layout they are in

`monorepo_components` decides *which job* the generator emits. The
generated job still has to tell the *scripts it runs* which directory
layout to expect — a standalone controller repo has the controller at
its root, while the monorepo has it at
`staging/services/$SERVICE-controller/`.

This is carried by a single environment variable, `MONO_REPO_ENABLED`:

- **Default `false`.** Every script behaves exactly as it does today.
  Local developer invocations and any not-yet-migrated standalone job are
  unaffected — this is the backward-compatibility guarantee.
- **Set to `true`** in the pod spec of every monorepo-variant job the
  generator emits. Scripts such as `build-controller.sh` then resolve the
  controller source as `$REPO_ROOT/staging/services/$SERVICE-controller`
  instead of `$REPO_ROOT/../$SERVICE-controller`; the code generator,
  release-tag, and test scripts resolve their paths accordingly.

The flag is **job-scoped and set by the generator** — derived from
`monorepo_components`, never exported as a global shell or CI-wide
variable. A globally exported `MONO_REPO_ENABLED=true` would force the
whole fleet into monorepo mode at once and defeat the incremental
migration; the default-`false`, generator-injected discipline is what
keeps a half-migrated fleet working.

A boolean is sufficient because the layout is conventional: the monorepo
is the Prow main ref, so its checkout path is the repo root, and the
existing `$AWS_SERVICE` variable identifies the subdirectory. No separate
"monorepo root path" variable is required.

## Prow job impact

The bulk of ACK's release machinery is unchanged because every post-release
job fires from the per-service repo, and the per-service repos still
receive pushes and tags (now via the publishing bot). The damage is
concentrated in tests and the post-release regeneration cascade.

### Unchanged

These all fire from per-service publish targets that still exist:

- `postsubmits/controller_release.tpl` — the entire post-tag chain
  (`{svc}-post-submit`, `{svc}-soak-on-release`,
  `{svc}-controller-release-tag`, `{svc}-controller-olm-bundle-pr`,
  `update-ack-chart`). All driven by per-service repo tags.
- `postsubmits/runtime_release.tpl`, `postsubmits/ack-chart_release.tpl`,
  `postsubmits/docs_website.tpl`, `postsubmits/test-infra.tpl`.
- All periodics that don't clone the in-scope source repos: label sync,
  CVE scan, EKS-distro upgrade probe, Go-version upgrade probe,
  lifecycle bot jobs.
- `presubmits/controller_bootstrap_test.tpl` (out of scope).

### Modified

These need to retarget from per-service repos to the monorepo, add
`run_if_changed` path filters, and reference monorepo subtrees instead of
sibling repos:

- `presubmits/service_tests.tpl` — main ref becomes
  `aws-controllers-k8s/ack`; each per-service block gets
  `run_if_changed: '^staging/services/$svc-controller/.*'`;
  `path_alias`/workdir scopes to the controller subtree.
- `presubmits/code_generator_tests.tpl` — integration tests resolve
  `code-generator`, `runtime`, and the target service from monorepo
  paths instead of separate clones.
- `presubmits/runtime_tests.tpl` — same pattern.
- `presubmits/pkg_tests.tpl` — main ref becomes the monorepo, scoped
  with `run_if_changed: '^staging/pkg/.*'`.
- `presubmits/test_infra_tests.tpl` — `extra_refs` paths for
  `code-generator` and the target service shift to monorepo subtrees.
- `postsubmits/community_docs.tpl` — trigger key renames `community` →
  `ack`; doc-build script paths shift accordingly.
- `postsubmits/controller_bootstrap_update.tpl` — today this fan-out
  clones every `$SVC-controller` to update static files. Under the new
  model those repos are bot-write-only; the job should land its changes
  in the monorepo instead and let the bot propagate.
- `periodics/docs_release.tpl` — collapses three clones (community,
  runtime, all services) into one monorepo clone.

### Obsolete

- `postsubmits/codegen_release.tpl` — the `auto-generate-controllers`
  cascade. Today it triggers on a `code-generator` tag, clones all 70
  services, and regenerates them. Under the monorepo model regeneration
  is a development-time activity in-tree: a `code-generator` change that
  affects services is committed alongside the regenerated service code in
  a single PR. The cascade has no remaining purpose and is removed.

### New

- **`postsubmits/ack_publish.tpl`** — postsubmit on `ack:main` that runs
  the upstream `kubernetes/publishing-bot` binary, scoped to rules whose
  source paths were touched in the merge. Primary mirror trigger.
- **`periodics/ack_publish_heartbeat.tpl`** — hourly cron running the
  bot across all rules as a backup, catching anything the postsubmit
  missed.
- **`presubmits/ack_validate_rules.tpl`** — validates `publishing/rules.yaml`
  against the actual monorepo layout. Every rule's source path must
  exist, and every directory under `staging/` must be covered by exactly
  one rule. Equivalent of upstream's `validate-rules`.
- **`presubmits/ack_workspace_build.tpl`** — `always_run: true`, runs
  `go build ./...` across the workspace. Catches breakage from top-level
  changes (e.g. `go.work` edits) that no service-scoped job would have
  caught.
- **`presubmits/ack_workspace_vet.tpl`** — `go vet ./...` across the
  workspace. Cheap monorepo-wide smoke.
- **`presubmits/ack_go_work_tidy.tpl`** — verifies `go.work` and per-module
  `go.sum` are tidy.
- **`postsubmits/library_release_tag.tpl`** — a `release-tag`-style
  postsubmit on the `runtime`, `code-generator`, and `pkg` publish
  targets. Reads the module's `VERSION` file, compares it to the latest
  Git tag, and tags the repo on a valid bump. See
  [Tagging library modules](#tagging-library-modules).
- **Auto-closer for downstream human PRs** — implemented as a GitHub
  Action (or small Prow plugin) on each `$SVC-controller`, `runtime`,
  `code-generator`, and `pkg` repo. On any PR opened by a non-bot
  account, post a redirect to the monorepo and close.

## Migration plan

A big-bang cutover is unnecessary and risky. The migration proceeds as a
sequence of PRs across `code-generator`, `test-infra`, and the monorepo.
Within a stage, PRs marked ∥ are independent and may land in parallel.

The import order is `pkg` → one pilot controller → `runtime` and
`code-generator` → remaining controllers. `pkg` goes first because it is a
dependency leaf and the simplest mirror target, so it de-risks the bot. A
single controller is migrated next — before the shared dependencies — to
prove the full controller release chain end to end early.

### Stage 1 — Rename

- **GitHub admin** (not a PR): rename `community` → `ack`. GitHub
  redirects preserve existing links, issues, stars, and clones. Done
  first so every subsequent monorepo PR targets the final repo name.

### Stage 2 — Backward-compatible tooling prep

No behavior change; the existing multi-repo flow keeps working. All ∥.

- **PR → `code-generator`**: teach `build-controller.sh` and the scripts
  it calls to honor `MONO_REPO_ENABLED`. When set, resolve the controller
  source as `staging/services/$SERVICE-controller` within the monorepo;
  when unset or `false`, keep today's `../$SERVICE-controller` behavior so
  the current per-repo flow is unbroken.
- **PR → `test-infra`**: add the `monorepo_components` list to
  `jobs_config.yaml` (initially empty) and generator support for emitting
  monorepo-variant jobs that set `MONO_REPO_ENABLED=true`. Generator
  output is unchanged while the list is empty. See
  [Configuration: tracking migrated components](#configuration-tracking-migrated-components).
- **PR → `test-infra`**: add the publishing-bot deployment manifests
  (Deployment + hourly CronJob, RBAC, rules mount) referencing the
  upstream `kubernetes/publishing-bot` image. Not yet wired to anything.
- **PR → `test-infra`**: add the new Prow job templates
  (`ack_workspace_build`/`vet`, `ack_go_work_tidy`, `ack_validate_rules`,
  `ack_publish`, `ack_publish_heartbeat`, `library_release_tag`,
  auto-closer). Dormant — they key on `ack` / `monorepo_components`
  membership, which nothing has yet.

### Stage 3 — Monorepo skeleton

- **PR → `ack`**: scaffold the repo — empty `go.work`, empty
  `publishing/rules.yaml`, the `staging/` and `staging/services/`
  directories, a top-level `Makefile`, updated `CONTRIBUTING.md` and
  `README.md`.

### Stage 4 — Import `pkg`, bring up the bot

- **PR → `ack`**: import `staging/pkg/` with full history
  (`git filter-repo`), add it to `go.work`, add `staging/pkg/VERSION`,
  add its `publishing/rules.yaml` entry.
- Enable the `ack_publish` postsubmit for `pkg` only. Verify the mirror
  to the `pkg` repo and the `library-release-tag` flow.
- Lock the `pkg` source repo to bot-write-only; enable its auto-closer.

### Stage 5 — Pilot one controller

- **PR → `ack`**: import the pilot controller into
  `staging/services/$svc-controller/` with full history. Add it to
  `go.work` and workspace-replace `pkg`. It continues to consume
  `runtime` as a tagged `go.mod` dependency for now (runtime is not yet
  in-tree). Add its `rules.yaml` entry.
- **PR → `test-infra`**: add `$svc` to `monorepo_components`, regenerate
  jobs.
- Validate end to end: open a release PR in `ack`, confirm a tag appears
  on `$SVC-controller`, confirm the existing image and chart publishing
  fires unchanged.
- Lock the `$SVC-controller` source repo to bot-write-only; enable its
  auto-closer.

`monorepo_components` membership is per-component, so during this stage
the pilot controller's presubmits run in monorepo mode for the controller
itself while still pulling `runtime` and `code-generator` as external
`extra_refs` — those do not move in-tree until Stage 6. The generator
emits this mixed form automatically by checking membership per
dependency; no special handling is needed.

### Stage 6 — Import `runtime` and `code-generator`

- **PR → `ack`**: import `staging/runtime/` with full history, add to
  `go.work`, workspace-replace `pkg`, add `staging/runtime/VERSION` and
  its `rules.yaml` entry.
- **PR → `ack`**: import `staging/code-generator/` with full history,
  same treatment.
- **PR → `ack`**: retrofit the Stage 5 pilot controller to
  workspace-replace `runtime` now that it is in-tree.
- Lock the `runtime` and `code-generator` source repos to
  bot-write-only; enable their auto-closers.

### Stage 7 — Wave migration

Migrate the remaining service controllers in batches of ~10. Per wave:

- **PR → `ack`**: import the batch into `staging/services/`, each with
  full history, a `go.work` entry, workspace-replace on `runtime`/`pkg`,
  and a `rules.yaml` entry.
- **PR → `test-infra`**: add the batch to `monorepo_components`,
  regenerate jobs.
- Lock each source repo to bot-write-only and enable its auto-closer
  after its wave.

### Stage 8 — Cleanup

- **PR → `test-infra`**: delete `codegen_release.tpl`; remove the
  standalone-mode job templates once the last service is flipped.
- **PR → `ack`**: finalize docs and onboarding.
- Archive (or leave bot-write-only) all source repos.

## Trade-offs and known costs

- **Per-service repos lose human PRs.** External contributors cannot PR
  against `$SERVICE-controller` repos anymore. They will be redirected to the
  monorepo. This is a deliberate centralization of contribution flow.
- **Monorepo size.** ~70 service controllers plus history will produce a
  large repository. Shallow clones, sparse checkouts, and judicious history
  rewriting at import time mitigate this but do not eliminate it.
- **Branch protection asymmetry.** GitHub branch protection is repo-wide;
  per-service teams cannot have admin scoped to their
  `staging/services/$SVC-controller/` subtree. CODEOWNERS gives
  path-scoped review but not path-scoped admin.
- **Audience overlap.** Renaming `community` → `ack` means the same repo
  hosts governance, docs, design proposals, *and* the source for every
  controller. This conflates two audiences (community/process vs.
  developer/code). An alternative considered (a fresh `aws-controllers-k8s/ack`
  repo, leaving `community` alone) trades GitHub redirects/history for
  cleaner audience separation. This proposal goes with the rename; it is
  reversible if the conflation proves painful.

## Tag and CI policy

- **No SemVer tags on the monorepo itself.** All SemVer tags live on
  publish targets (`$SERVICE-controller`, `runtime`, `code-generator`,
  `pkg`). The monorepo carries no `v*.*.*` tags of its own, avoiding
  collision with what the publishing bot writes downstream.
- **CI gating on cross-cutting changes.** PRs that touch
  `staging/runtime/` or `staging/code-generator/` continue to use the
  existing
  `CodegenPresubmitServices` subset (~18 representative services) for e2e
  presubmit; the full fleet runs on a postsubmit or scheduled basis.
  Service-only PRs run e2e for the touched service.

## Tagging library modules

Service controllers are tagged off `helm/values.yaml` — the
`controller-release-tag` postsubmit reads `image.tag`, compares it to the
latest Git tag, and tags the per-service repo on a valid bump. The library
modules — `pkg`, `runtime`, `code-generator` — have no Helm chart and so
need an equivalent driver.

We propose a **`VERSION` file in each library module directory**:

```
ack/
  staging/pkg/VERSION             # e.g. v0.3.1
  staging/runtime/VERSION         # e.g. v0.45.0
  staging/code-generator/VERSION  # e.g. v0.45.0
```

The `VERSION` file is the single source of truth for that module's next
release. It is a plain file inside the module subtree, so the publishing
bot mirrors it into the publish target like any other file — no special
handling, no bot-side injection.

We deliberately use a per-module `VERSION` file rather than a single
central `versions.yaml` because:

- It travels *with* the mirrored content. The downstream tag job reads it
  locally from the publish target checkout, exactly as
  `controller-release-tag` reads `helm/values.yaml` today.
- A central file would have to be injected into each publish target
  separately by the bot — more machinery, more drift surface.
- It keeps libraries and controllers on the same mental model: one
  in-tree file declares the next version; a postsubmit acts on it.

## Releasing a library module

The flow mirrors the controller release flow:

```
[ developer ]
     |
     |  PR against ack/staging/runtime/VERSION (v0.45.0 -> v0.46.0)
     v
[ ack monorepo: PR merges to main ]
     |
     |  publishing-bot mirrors the commit onto runtime:main
     v
[ runtime:main push ]
     |
     |  Prow postsubmit `library-release-tag` fires,
     |  reads staging/runtime/VERSION, compares to git describe,
     |  validates next-patch-or-minor, pushes tag v0.46.0
     v
[ runtime: v0.46.0 tag ]
     |
     |--> existing runtime release/docs jobs fire as today
```

A new `postsubmits/library_release_tag.tpl` provides this job for the
`runtime`, `code-generator`, and `pkg` publish targets. It is a
generalization of `controller-release-tag.sh`: read `VERSION` when no
`helm/values.yaml` is present, otherwise behave identically (compare to
`git describe`, accept only the next patch or minor, tag and push, open a
tracking issue on failure).

Because `runtime` is consumed in-tree via the workspace, a `VERSION` bump
does not by itself change what services build against during development.
Services pick up a new `runtime` release when the publishing bot rewrites
their published `go.mod` to the newly tagged version — the `rules.yaml`
entry for each service pins which `runtime` tag is used downstream.
