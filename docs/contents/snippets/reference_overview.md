{% macro render_service(name, resources) %}
## {{ name }}

Resource | API Version
:--------|:-----------
{% for resource in resources | sort(attribute='name') -%}
[{{ resource.name }}](../{{ name }}/{{ resource.apiVersion }}/{{ resource.name }}) | {{ resource.apiVersion }}
{% endfor %}
{% endmacro %}

{% for service in page.meta.services | sort(attribute='name') %}
{{ render_service(service.name, service.resources) }}{% endfor %}