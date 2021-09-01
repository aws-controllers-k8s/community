var docs = [
{{ range $index, $page := (where .Site.Pages "Section" "in" .Site.Params.documentationSections) -}}
  {
    id: {{ $index }},
    title: "{{ .Title }}",
    description: "{{ .Params.description }}",
    href: "{{ .URL | relURL }}"
  },
{{ end -}}
];