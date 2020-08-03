#!/usr/bin/env bash

ensure_controller_gen() {
  if ! is_installed controller-gen; then
    # Need this version of controller-gen to allow dangerous types and float
    # type support
    go get sigs.k8s.io/controller-tools/cmd/controller-gen@4a903ddb7005459a7baf4777c67244a74c91083d
  fi
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
