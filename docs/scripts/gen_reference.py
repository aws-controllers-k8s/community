#!/usr/bin/env python3.8

from __future__ import annotations
import os

import yaml
import argparse

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
    res_spec = []
    res_status = []

    names = ResourceNames(
        kind=crd['spec']['names']['kind'],
        listKind=crd['spec']['names']['listKind'],
        plural=crd['spec']['names']['plural'],
        singular=crd['spec']['names']['singular']
    )

    spec = schema['properties']['spec']
    spec_properties = spec['properties']
    for field_name, field in spec_properties.items():
        required  = field_name in spec['required'] if 'required' in spec else False
        res_spec.append(convert_field(field_name, field, required))

    status = schema['properties']['status']
    status_properties = status['properties']
    for field_name, field in status_properties.items():
        required  = field_name in status['required'] if 'required' in status else False
        res_status.append(convert_field(field_name, field, required))

    return Resource(
        name=crd['spec']['names']['kind'],
        service=service, #TODO(RedbackThomson): Load from file path?
        group=crd['spec']['group'],
        apiVersion=ver['name'],
        description=schema['description'],
        scope=crd['spec']['scope'],
        names=names,
        spec=res_spec,
        status=res_status
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

def main(gopath: Path, go_src_parent: Path, service_bases_path: Path, output_path: Path):
    overviews = []

    src_parent = gopath / go_src_parent

    for controller_repo in src_parent.glob("*-controller"):
        if not controller_repo.is_dir() or \
            controller_repo.stem.startswith("template"):
            continue

        service_bases_path = controller_repo / bases_path

        if not service_bases_path.exists():
            raise ValueError(f"Service base path {service_bases_path} does not exist")

        # Get service name from repository
        # TODO(RedbackThomson): Find an elegant way to get name-cased 
        service = controller_repo.stem[:-len("-controller")]

        resources = write_service_pages(service, service_bases_path, output_path)
        overviews.append(ServiceOverview(service, resources))

    write_overview_page(overviews, output_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate documentation reference pages")
    parser.add_argument("--bases_path", type=str, default="./config/crd/bases",
        help="Relative path to the `bases` directory, relative to the service controller root")
    parser.add_argument("--gopath", type=str, default=os.environ.get('GOPATH'),
        help="Path of the GOPATH - Defaults to $GOPATH")
    parser.add_argument("--go_src_parent", type=str, default="./src/github.com/aws-controllers-k8s",
        help="Relative path to the ACK src path, relative to the GOPATH")
    parser.add_argument("--output_path", type=str, default="./contents/reference",
        help="Relative path to the documentation output directory, relative to the `docs` directory")

    args = parser.parse_args()

    bases_path = Path(args.bases_path)
    gopath = Path(args.gopath)
    go_src_parent = Path(args.go_src_parent)
    output_path = Path(args.output_path)

    main(gopath, go_src_parent, bases_path, output_path)