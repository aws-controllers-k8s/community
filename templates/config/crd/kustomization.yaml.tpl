apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
{{- range .CRDNames }}
  - bases/{{ $.APIGroup }}_{{ . }}.yaml 
{{- end }}
