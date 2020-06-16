# Build the manager binary
FROM golang:1.14.1 as builder

ARG work_dir=/github.com/aws/aws-service-operator-k8s
WORKDIR $work_dir
# For building Go Module required
ENV GOPROXY=direct
ENV GO111MODULE=on
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  go mod download
# Copy the go source
# TODO: for now copy whole repository. Later once autogeneration is ready, need to determine which dependencies to take in
COPY . $work_dir/
# Build
RUN  go build -a -o $work_dir/bin/manager $work_dir/services/bookstore/cmd/controller/main.go


FROM amazonlinux:2
ARG work_dir=/github.com/aws/aws-service-operator-k8s
WORKDIR /
COPY --from=builder $work_dir/bin/manager /bin/.
ENTRYPOINT ["/bin/manager"]
