#!/usr/bin/env bash

ensure_controller_gen() {
  go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0
}

ensure_service_controller_running() {
  echo "Ensure that service controllers are running for given version"
  # Give executable permission
  chmod +x ./scripts/generate-crds.sh

  __image_version=$1
  __crd_path=$2

  for d in ./services/*; do
    if [ -d "$d" ]; then
      echo "service: $d"
      ./scripts/generate-crds.sh "$d" "$__crd_path"
        for f in "$d"/*; do
          if [ "$f" = "$d"/"Dockerfile" ]; then
            echo "$f"
            __service_name=$(basename "$d")
            __ack_image_tag="$__service_name"-"$__image_version"
            ensure_ecr_image "$__ack_image_tag" "$f"
          fi
        done
    fi
  done

  kubectl apply -f "$__crd_path"

  echo "Installing/Upgrading helm chart"
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