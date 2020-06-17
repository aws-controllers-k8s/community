#!/usr/bin/env bash

ensure_crd_gen() {
    local __crd_dir="$1"
    CONTROLLER_GEN=$(which controller-gen)
    if [[ -z "$CONTROLLER_GEN"  ]]
    then
      set -e
      CONTROLLER_GEN_TMP_DIR="$(mktemp -d)"
      cd "$CONTROLLER_GEN_TMP_DIR"
      go mod init tmp
      go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0
      rm -rf "$CONTROLLER_GEN_TMP_DIR"
      CONTROLLER_GEN=$(GOBIN)/controller-gen
    fi
    $CONTROLLER_GEN "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config="$__crd_dir"
}
