#!/usr/bin/env bash

# A script that builds Helm packages for all service controllers that have a
# Helm chart, builds the Helm repository index.yaml and git commits and pushes
# the updated Helm packages and repository index to the gh-pages branch of the
# upstream source repository.

set -euxo pipefail

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
ROOT_DIR="$SCRIPTS_DIR/.."
BUILD_DIR="$ROOT_DIR/build"
SERVICES_DIR="$ROOT_DIR/services"
DEFAULT_HELM_REPO_URL="https://aws.github.io/aws-controllers-k8s/charts"
DEFAULT_GH_USER_EMAIL="eks-bot@users.noreply.github.com"
DEFAULT_GH_USER_EMAIL="eks-bot"
DEFAULT_GIT_REPOSITORY="https://eks-bot:${GITHUB_TOKEN}@github.com/aws/aws-controllers-k8s.git"
VERSION=$(git describe --tags --always --dirty || echo "unknown")

: "${HELM_REPO_URL:=$DEFAULT_HELM_REPO_URL}"
: "${GH_USER_NAME:=$DEFAULT_GH_USER_NAME}"
: "${GH_USER_EMAIL:=$DEFAULT_GH_USER_EMAIL}"
: "${GIT_REPOSITORY:=$DEFAULT_GIT_REPOSITORY}"
: "${GIT_COMMIT:="false"}"

source "$SCRIPTS_DIR/lib/common.sh"

check_is_installed helm "You can install Helm with the helper scripts/install-helm.sh"

USAGE="
Usage:
  $(basename "$0")

Environment variables:
  HELM_REPO_URL:            The URL for the Helm repository.
                            Default: $DEFAULT_HELM_REPO_URL
  GH_USER_NAME:             The name of the Github user to use when Git
                            commit'ing.
                            Default: $DEFAULT_GH_USER_NAME
  GH_USER_EMAIL:            The email of the Github user to use when Git
                            commit'ing.
                            Default: $DEFAULT_GH_USER_EMAIL
  GIT_REPOSITORY:           The Git repository URL to commit to.
                            Default: $DEFAULT_GIT_REPOSITORY
  GIT_COMMIT:               If false (default), only build the packages and
                            index. If true, also creates a Git commit and
                            pushes that commit to an upstream Git repository.
                            Default: false
"

CHARTS_DIR=$ROOT_DIR/charts

if [[ $GIT_COMMIT = "false" ]]; then
    # On a dry run, we stash the charts in the git-ignored build/ directory.
    # For non-dry-run, we use the $ROOT_DIR/charts directory, which in the
    # gh-pages branch of the aws/aws-controllers-k8s upstream source repository
    # contains the chart packages and index.yaml file.
    CHARTS_DIR=$BUILD_DIR/charts
fi

mkdir -p $CHARTS_DIR

export PACKAGE_OUTPUT_PATH="$BUILD_DIR/release/$VERSION"

for SERVICE_DIR in $SERVICES_DIR/*; do
    SERVICE=$( basename $SERVICE_DIR)
    if [[ -d "$SERVICES_DIR/$SERVICE/helm" ]]; then
        echo -n "Generating Helm chart package for $SERVICE ... "
        $SCRIPTS_DIR/helm-package-controller.sh $SERVICE 1>/dev/null || exit 1
        echo "ok."
    fi
done

# We need to place the packages into the Helm repository's root directory.
mv -f $PACKAGE_OUTPUT_PATH/*.tgz $CHARTS_DIR

echo -n "Building index for Helm repo ... "
helm repo index $CHARTS_DIR --url $HELM_REPO_URL 1>/dev/null || exit 1
echo "ok."

if [[ $GIT_COMMIT = "true" ]]; then
    git config user.name $GH_USER_NAME
    git config user.email $GH_USER_EMAIL
    git remote set-url upstream $GIT_REPOSITORY
    git checkout gh-pages
    git add .
    git commit -m "Publish ACK service controller charts for $VERSION"
    git push upstream gh-pages
fi
