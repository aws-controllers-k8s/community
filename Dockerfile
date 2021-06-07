# Base image to use for the final stage
ARG base_image=amazonlinux:2
# Build the manager binary
FROM golang:1.14.1 as builder

ARG service_alias
# The tuple of controller image version information
ARG service_controller_git_version
ARG service_controller_git_commit
ARG build_date
# The directory within the builder container into which we will copy our
# service controller code.
ARG work_dir=/github.com/aws-controllers-k8s/$service_alias-controller
WORKDIR $work_dir
# For building Go Module required
ENV GOPROXY=https://proxy.golang.org,direct
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV VERSION_PKG=github.com/aws-controllers-k8s/$service_alias-controller/pkg/version
# Copy the Go Modules manifests and LICENSE/ATTRIBUTION
COPY $service_alias-controller/LICENSE $work_dir/LICENSE
COPY $service_alias-controller/ATTRIBUTION.md $work_dir/ATTRIBUTION.md
# Copy Go mod files
COPY $service_alias-controller/go.mod $work_dir/go.mod
COPY $service_alias-controller/go.sum $work_dir/go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Now copy the go source code for the controller...
COPY $service_alias-controller/apis $work_dir/apis
COPY $service_alias-controller/cmd $work_dir/cmd
COPY $service_alias-controller/pkg $work_dir/pkg
# Build
RUN GIT_VERSION=$service_controller_git_version && \
    GIT_COMMIT=$service_controller_git_commit && \
    BUILD_DATE=$build_date && \
    go build -ldflags="-X ${VERSION_PKG}.GitVersion=${GIT_VERSION} \
    -X ${VERSION_PKG}.GitCommit=${GIT_COMMIT} \
    -X ${VERSION_PKG}.BuildDate=${BUILD_DATE}" \
    -a -o $work_dir/bin/controller $work_dir/cmd/controller/main.go

FROM $base_image
ARG service_alias
ARG work_dir=/github.com/aws-controllers-k8s/$service_alias-controller
WORKDIR /
COPY --from=builder $work_dir/bin/controller $work_dir/LICENSE $work_dir/ATTRIBUTION.md /bin/
ENTRYPOINT ["/bin/controller"]
