---
name: AWS service controller
about: Template for a new AWS service controller
title: "[name] service controller"
labels: Service Controller
assignees: ''

---

## ACK Service Controller
Support for [service name]

### Checklist for Dev Preview
- [ ] Code generation (`make build-controller`)
- [ ] End-to-end test (`make kind-test`)
- [ ] Docs are updated, at least:
  - [ ] [testing](https://aws.github.io/aws-controllers-k8s/dev-docs/testing/) 
  - [ ] [svc controller listing](https://aws.github.io/aws-controllers-k8s/services/)
- [ ] Container image in public registry
