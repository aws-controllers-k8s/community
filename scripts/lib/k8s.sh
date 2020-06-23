#!/usr/bin/env bash

ensure_controller_gen() {
  if ! is_installed controller-gen; then
    go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0
  fi
}

ensure_service_controller_running() {

  __service_path=$1
  __service=$2
  __image_version=$3
  __crd_path=$4

  echo "Ensuring that service controller $__service are running for given version"

  ./scripts/generate-crds.sh "$__service_path" "$__crd_path"
  for f in "$__service_path"/*; do
    if [ "$f" = "$__service_path/Dockerfile" ]; then
      __ack_image_tag="$__service"-"$__image_version"
      ensure_ecr_image "$__ack_image_tag" "$f"
    fi
  done

  kubectl apply -f "$__crd_path"

  echo "Installing/Upgrading helm chart"
  #TODO: Update it to deploy for one service
  ensure_helm_chart_installed "$__image_version"
  # Wait between two tests for old controllers to be replaced.
  # Using kubectl wait is a little tricky for this terminating condition,
  # as there are race condition if controller is deleted before wait command.
  # Using a sleep here keeps things simple and allows time for old controllers to flush out.
  # if there are any issues, ensuring new controller pods later will catch those problems.
  echo "Waiting for 120 seconds for old controllers to be terminated."
  sleep 120
  ensure_controller_pods

}