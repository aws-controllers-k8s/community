#!/usr/bin/env python3.8

from __future__ import annotations

import yaml

from pathlib import Path
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional, Union


@dataclass(frozen=True)
class ResourceNames:
    kind: str
    listKind: str
    plural: str
    singular: str

@dataclass(frozen=True)
class Field:
    name: str
    description: str
    type: str
    required: bool
    contains: Optional[Union[str, List[Field]]]
    contains_description: Optional[str]

@dataclass(frozen=True)
class Resource:
    name: str
    service: str
    group: str
    apiVersion: str
    description: str
    scope: str
    names: ResourceNames
    spec: List[Field]
    status: List[Field]

    def output_path(self) -> Path:
        return Path(f"{self.service}/{self.apiVersion}/{self.name}.md")

@dataclass(frozen=True)
class ResourcePageMeta:
    resource: Resource


@dataclass(frozen=True)
class ResourceOverview:
    name: str
    apiVersion: str

@dataclass(frozen=True)
class ServiceOverview:
    name: str
    resources: ResourceOverview

@dataclass(frozen=True)
class OverviewPageMeta:
    services: List[ServiceOverview]

def load_crd_yaml(path: str) -> Dict:
    with open(path, "r") as stream:
        return yaml.safe_load(stream)

def convert_field(name: str, property: Dict, required=False) -> Field:
    field_type = property['type']
    contains = None
    contains_description = None
    if field_type == "array":
        contains_description = property['items']['description'] if 'description' in property['items'] else ''
        contains = None
        # Check for array of objects
        if 'properties' in property['items']:
            contains = []
            for subfield_name, subfield in property['items']['properties'].items():
                subfield_required  = subfield_name in property['items']['required'] if 'required' in property else False
                contains.append(convert_field(subfield_name, subfield, subfield_required))
        else:
            contains = property['items']['type']

    elif field_type == "object":
        contains = []
        # TODO (RedbackThomson): Handle arbitrary maps
        if 'additionalProperties' in property:
            contains = property['additionalProperties']['type']
        else:
            for subfield_name, subfield in property['properties'].items():
                subfield_required = subfield_name in property['required'] if 'required' in property else False
                contains.append(convert_field(subfield_name, subfield, subfield_required))

    return Field(
        name=name,
        description=property['description'] if 'description' in property else '',
        type=property['type'],
        required=required,
        contains=contains,
        contains_description=contains_description
    )

def convert_crd_to_resource(crd: Dict, service) -> Resource:
    ver = crd['spec']['versions'][0]
    schema = ver['schema']['openAPIV3Schema']
    spec = []
    status = []

    names = ResourceNames(
        kind=crd['spec']['names']['kind'],
        listKind=crd['spec']['names']['listKind'],
        plural=crd['spec']['names']['plural'],
        singular=crd['spec']['names']['singular']
    )

    spec_properties = schema['properties']['spec']['properties']
    for field_name, field in spec_properties.items():
        required  = field_name in schema['properties']['spec']['required']
        spec.append(convert_field(field_name, field, required))

    status_properties = schema['properties']['status']['properties']
    for field_name, field in status_properties.items():
        required  = field_name in schema['properties']['status']['required']
        status.append(convert_field(field_name, field, required))

    return Resource(
        name=crd['spec']['names']['kind'],
        service=service, #TODO(RedbackThomson): Load from file path?
        group=crd['spec']['group'],
        apiVersion=ver['name'],
        description=schema['description'],
        scope=crd['spec']['scope'],
        names=names,
        spec=spec,
        status=status
    )

def write_service_pages(service: str, service_path: Path, output_path: Path) -> List[ResourceOverview]:
    resources = []

    for crd_path in Path(service_path).rglob('*.yaml'):
        crd = load_crd_yaml(crd_path)

        resource = convert_crd_to_resource(crd, service)
        resources.append(ResourceOverview(name=resource.name,
            apiVersion=resource.apiVersion))

        resource_path = output_path / resource.output_path()
        resource_path.parent.mkdir(parents=True, exist_ok=True)

        with open(resource_path, "w") as out:
            page_meta = ResourcePageMeta(resource=resource)
            # TODO(RedbackThomson): Clean up templating
            print("---", file=out)
            yaml.dump(asdict(page_meta), out)
            print("---", file=out)
            print('{% include "reference.md" %}', file=out)
    
    return resources

def write_overview_page(services: List[ServiceOverview], output_path: Path):
    overview_path = output_path / "overview.md"

    with open(overview_path, "w") as out:
        page_meta = OverviewPageMeta(services=services)
        # TODO(RedbackThomson): Clean up templating
        print("---", file=out)
        yaml.dump(asdict(page_meta), out)
        print("---", file=out)
        print('{% include "reference_overview.md" %}', file=out)

def main(base_directory: Path, output_path: Path):
    overviews = []

    for base_path in base_directory.iterdir():
        if not base_path.is_dir():
            continue

        service = base_path.stem
        resources = write_service_pages(service, base_path, output_path)
        overviews.append(ServiceOverview(service, resources))

    write_overview_page(overviews, output_path)


if __name__ == "__main__":
    crd_path = Path("./scripts/bases")
    output_path = Path("./contents/reference")
    main(crd_path, output_path)