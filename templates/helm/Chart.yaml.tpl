apiVersion: v1
name: ack-{{ .ServiceIDClean }}-controller
description: A Helm chart for the ACK service controller for {{ .ServiceIDClean }}
version: {{ .ReleaseVersion }}
appVersion: {{ .ReleaseVersion }}
home: https://github.com/aws/aws-controllers-k8s
icon: https://raw.githubusercontent.com/aws/eks-charts/master/docs/logo/aws.png
sources:
  - https://github.com/aws/aws-controllers-k8s
maintainers:
  - name: ACK Admins
    url: https://github.com/orgs/aws/teams/aws-controllers-for-kubernetes-ack-admins
keywords:
  - aws
  - kubernetes
  - {{ .ServiceIDClean }}
