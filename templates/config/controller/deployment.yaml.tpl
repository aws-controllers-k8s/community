apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller
  name: ack-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ack-{{ .ServiceAlias}}-controller
  namespace: ack-system
  labels:
    control-plane: controller
spec:
  selector:
    matchLabels:
      control-plane: controller
  replicas: 1
  template:
    metadata:
      labels:
        control-plane: controller
    spec:
      containers:
      - command:
        - ./bin/controller
        args:
        # Obviously this needs to change...
        - --aws-account-id
        - "123456"
        image: controller:latest
        name: controller
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 200Mi
      terminationGracePeriodSeconds: 10
