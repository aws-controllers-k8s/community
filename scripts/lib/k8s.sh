#!/usr/bin/env bash

CONTROLLER_TOOLS_VERSION="v0.4.0"

# ensure_controller_gen checks that the `controller-gen` binary is available on
# the host system and if it is, that it matches the exact version that we
# require in order to standardize the YAML manifests for CRDs and Kubernetes
# Roles.
#
# If the locally-installed controller-gen does not match the required version,
# prints an error message asking the user to uninstall it.
#
# NOTE: We use this technique of building using `go build` within a temp
# directory because controller-tools does not have a binary release artifact
# for controller-gen.
#
# See: https://github.com/kubernetes-sigs/controller-tools/issues/500
ensure_controller_gen() {
    if ! is_installed controller-gen; then
        # GOBIN not always set... so default to installing into $GOPATH/bin if
        # not...
        __install_dir=${GOBIN:-$GOPATH/bin}
        __install_path="$__install_dir/controller-gen"
        __work_dir=$(mktemp -d /tmp/controller-gen-XXX)

        cd "$__work_dir"

        go mod init tmp
        go get -d "sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_TOOLS_VERSION}"
        go build -o "$__work_dir/controller-gen" sigs.k8s.io/controller-tools/cmd/controller-gen
        mv "$__work_dir/controller-gen" "$__install_path"

        rm -rf "$WORK_DIR"
        echo "****************************************************************************"
        echo "WARNING: You may need to reload your Bash shell and path. If you see an"
        echo "         error like this following:"
        echo ""
        echo "Error: couldn't find github.com/aws/aws-sdk-go in the go.mod require block"
        echo ""
        echo "simply reload your Bash shell with \`exec bash\`" and then re-run whichever
        echo "command you were running."
        echo "****************************************************************************"
    else
        # Don't overide the existing version let the user decide.
        if ! controller_gen_version_equals "$CONTROLLER_TOOLS_VERSION"; then
            echo "FAIL: Existing version of controller-gen "`controller-gen --version`", required version is $CONTROLLER_TOOLS_VERSION."
            echo "FAIL: Please uninstall controller-gen and re-run this script, which will install the required version."
            exit 1
        fi
    fi
}

# controller_gen_version_equals accepts a string version and returns 0 if the
# installed version of controller-gen matches the supplied version, otherwise
# returns 1
#
# Usage:
#
#   if controller_gen_version_equals "v0.4.0"; then
#       echo "controller-gen is at version 0.4.0"
#   fi
controller_gen_version_equals() {
    currentver="$(controller-gen --version | cut -d' ' -f2 | tr -d '\n')";
    requiredver="$1";
    if [ "$currentver" = "$requiredver" ]; then
        return 0
    else
        return 1
    fi;
}

# ensure_kubectl installs the kubectl binary if it isn't present on the system.
# It installs the kubectl binary for the latest stable release of Kubernetes
# and uses `sudo mv` to place the downloaded binary into your PATH.
ensure_kubectl() {
    if ! is_installed kubectl ; then
        __platform=$(uname | tr '[:upper:]' '[:lower:]')
        __stable_k8s_version=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
        curl -Lo "https://storage.googleapis.com/kubernetes-release/release/$__stable_k8s_version/bin/$__platform/amd64/kubectl"
        chmod +x kubectl
        sudo mv ./kubectl /usr/local/bin/kubectl
    fi
}

# ensure_kustomize installs the kustomize binary if it isn't present on the
# system and uses `sudo mv` to place the downloaded binary into your PATH.
ensure_kustomize() {
    if ! is_installed kustomize ; then
        curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
        chmod +x kustomize
        sudo mv kustomize /usr/local/bin/kustomize
    fi
}

ensure_service_controller_running() {

  __service_path=$1
  __service=$2
  __image_version=$3
  __crd_path=$4

  echo "Ensuring that service controller $__service are running for given version"

  ./scripts/generate-crds.sh "$__service_path" "$__crd_path"
  kubectl apply -f "$__crd_path"

  for f in "$__service_path"/*; do
    if [ "$f" = "$__service_path/Dockerfile" ]; then
      __ack_image_tag="ack-$__service-$__image_version"
      ensure_ecr_image "$__ack_image_tag" "$f"

      echo "Installing/Upgrading helm chart"
      ensure_helm_chart_installed "$__service" "$__ack_image_tag"
      # Wait between two tests for old controllers to be replaced.
      # Using kubectl wait is a little tricky for this terminating condition,
      # as there are race condition if controller is deleted before wait command.
      # Using a sleep here keeps things simple and allows time for old controllers to flush out.
      # if there are any issues, ensuring new controller pods later will catch those problems.
      echo "Waiting for 120 seconds for old controllers to be terminated."
      sleep 120
      ensure_controller_pods
      break
    fi
  done
}

# resource_exists returns 0 when the supplied resource can be found, 1
# otherwise. An optional second parameter overrides the Kubernetes namespace
# argument
k8s_resource_exists() {
    local __res_name=${1:-}
    local __namespace=${2:-}
    local __args=""
    if [ -n "$__namespace" ]; then
        __args="$__args-n $__namespace"
    fi
    kubectl get $__args "$__res_name" >/dev/null 2>&1
}

# get_field_from_status returns the field from status of a K8s resource
# get_field_from_status accepts three parameters. namespace (which is an optional parameter),
# resource_name and status_field
get_field_from_status() {

  if [[ "$#" -lt 2 || "$#" -gt 3 ]]; then
    echo "[FAIL] Usage: get_field_from_status [namespace] resource_name status_field"
    exit 1
  fi

  local __namespace=""
  local __resource_name=""
  local __status_field=""

  if [[ "$#" -eq 2 ]]; then
    __resource_name="$1"
    __status_field="$2"
  else
    __namespace="$1"
    __resource_name="$2"
    __status_field="$3"
  fi

  local __args=""
  if [ -n "$__namespace" ]; then
      __args="$__args-n $__namespace"
  fi


  local __id=$(kubectl get $__args "$__resource_name" -o=json | jq -r .status."$__status_field")
  if [[ -z "$__id" ]];then
    echo "FAIL: $__resource_name resource's status does not have $__status_field field"
    exit 1
  fi
  echo "$__id"
}
