# {{ page.meta.resource.name }} 
`{{ page.meta.resource.group }}/{{ page.meta.resource.apiVersion }}`

Property|Value
--------|-----
Scope|{{ page.meta.resource.scope }}
Kind|`{{ page.meta.resource.names.kind }}`
ListKind|`{{ page.meta.resource.names.listKind }}`
Plural|`{{ page.meta.resource.names.plural }}`
Singular|`{{ page.meta.resource.names.singular }}`

{{ page.meta.resource.description }}

## Spec

{% macro render_yaml_field(field, indentation=0, has_list_delim=False) -%}
{%- if field.type == "array" -%}{{ field.name }}:
{% if field.contains is not string -%}
{%- for subfield in field.contains -%}
{{ render_yaml_field(subfield, indentation + 1 + (-1 if loop.first else 0), loop.first) | indent((indentation + 1) * 2 + (-2 if loop.first else 0), True) }}
{% endfor -%}
{%- else -%}
- {{ field.contains }}
{%- endif -%}
{%- elif field.type == "object" -%}
{{ field.name }}:{%- if field.contains is string %} {}{%- else %}
{% for subfield in field.contains -%}
{{ render_yaml_field(subfield, indentation + 1, false) | indent((indentation + 1) * 2, True) }}
{% endfor -%}
{%- endif -%}
{%- else -%}
{{ "- " if has_list_delim }}{{ field.name }}: {{ field.type }}
{%- endif -%}
{%- endmacro %}

```yaml
{% for field in page.meta.resource.spec %}
{{ render_yaml_field(field) }}{% endfor %}
```

{% macro render_field(field, prefix='') -%}
<tr>
    <td><strong>{{ prefix }}{{ field.name }}</strong><br/>{{ "Required" if field.required else "Optional" }}</td>
    <td><strong>{{ field.type }}</strong><br/>{{ field.description }}</td>
</tr>
{% if field.type == "array" %}
    <tr>
        <td><strong>{{ prefix }}{{ field.name }}.[]</strong><br/>Required</td>
        <td><strong>{% if field.contains is not string %}object{% else %}{{ field.contains }}{% endif %}</strong><br/>{{ field.contains_description if field.contains_description else "" }}</td>
    </tr>
    {% if field.contains is not string %}
        {% for subfield in field.contains %}
            {{ render_field(subfield, prefix + field.name + ".[].") }}
        {% endfor %}
    {% endif %}
{% elif field.type == "object" %}
    {% if field.contains is not string %}
        {% for subfield in field.contains %}
            {{ render_field(subfield, prefix + field.name + ".") }}
        {% endfor %}
    {% endif %}
{% endif %}
{%- endmacro %}


<table>
    <tr>
        <th colspan="2">Fields</th>
    </tr>
    {% for field in page.meta.resource.spec %}
        {{ render_field(field) }}
    {% endfor %}
</table>

## Status

```yaml
{% for field in page.meta.resource.status %}
{{ render_yaml_field(field) }}{% endfor %}
```

<table>
    <tr>
        <th colspan="2">Fields</th>
    </tr>
    {% for field in page.meta.resource.status %}
        {{ render_field(field) }}
    {% endfor %}
</table>