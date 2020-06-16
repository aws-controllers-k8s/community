# Build the manager binary
FROM golang:1.14.1 as builder

ARG workDirName=/github.com/aws/aws-service-operator-k8s
WORKDIR $workDirName
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
COPY . $workDirName/
# Build
RUN  go build -a -o bin/manager $workDirName/services/bookstore/cmd/controller/main.go


FROM amazonlinux:latest
ARG workDirName=/github.com/aws/aws-service-operator-k8s
WORKDIR /
COPY --from=builder $workDirName/bin/manager bin/.
ENTRYPOINT ["./bin/manager"]
