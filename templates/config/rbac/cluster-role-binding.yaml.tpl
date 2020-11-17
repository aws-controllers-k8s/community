apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ack-{{ .ServiceIDClean }}-controller-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ack-{{ .ServiceIDClean }}-controller
subjects:
- kind: ServiceAccount
  name: default
  namespace: ack-system
